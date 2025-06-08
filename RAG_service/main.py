from fastapi import FastAPI, HTTPException
from fastapi import FastAPI
from contextlib import asynccontextmanager
from pydantic import BaseModel
from typing import List
from pathlib import Path
from tqdm import tqdm
import os
import numpy as np
import faiss


import boto3
from tempfile import TemporaryDirectory
from langchain_community.docstore.in_memory import InMemoryDocstore

from pymongo import MongoClient
from langchain_community.vectorstores import FAISS
from langchain_huggingface import HuggingFaceEmbeddings
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_community.document_loaders import TextLoader, PyPDFLoader
from langchain.docstore.document import Document
from langchain_community.vectorstores import FAISS

from threading import Lock
import threading
import time

import logging

# Настройка базового логгера
logging.basicConfig(
    level=logging.INFO,  # Уровень логирования (DEBUG, INFO, WARNING, ERROR, CRITICAL)
    format='%(asctime)s - %(levelname)s - %(message)s',  # Формат сообщения
    datefmt='%Y-%m-%d %H:%M:%S'  # Формат времени
)



FAISS_REBUILD_INTERVAL = 300  # секунд (например, 5 минут)


S3_BUCKET = "materials"
S3_PREFIX = "materials"

S3_CONFIG = {
    'aws_access_key_id': os.getenv("MINIO_ROOT_USER"),
    'aws_secret_access_key': os.getenv("MINIO_ROOT_PASSWORD"),
    'endpoint_url': os.getenv("S3_ENDPOINT_URL")  # можно использовать кастомный URL для S3-совместимого хранилища
}

# --- Конфигурация ---
INDEX_DIR = Path("./vectorstore/faiss_index")
EMBEDDING_MODEL_NAME = "intfloat/e5-base-v2"


MONGO_URI = os.getenv("MONGO_DB_URL")
MONGO_DB = os.getenv("MONGO_DB")
MONGO_COLLECTION = os.getenv("MONGO_COLLECTION")

CHUNK_SIZE = 800
CHUNK_OVERLAP = 50

# --- FastAPI ---

@asynccontextmanager
async def lifespan(app: FastAPI):
    init_vectorstore()
    start_faiss_rebuilder()
    yield  # здесь запускается сервер
    # Здесь можно закрыть ресурсы при завершении

app = FastAPI(lifespan=lifespan, title="RAG FastAPI with MongoDB")


# --- MongoDB ---
mongo_client = MongoClient(MONGO_URI)
db = mongo_client[MONGO_DB]
collection = db[MONGO_COLLECTION]

# --- Модель эмбеддингов ---
embedding_model = HuggingFaceEmbeddings(model_name=EMBEDDING_MODEL_NAME)

# --- Инициализация FAISS ---
vectorstore = None


# --- Подготовка документов ---
def clean_text(text: str) -> str:
    text = text.replace("-\n", "").replace("\n", " ")
    return text


def sync_mongo_with_s3(bucket: str, prefix: str, s3_config: dict) -> bool:
    """
    Синхронизирует Mongo и S3:
    - Удаляет из Mongo чанки, если исходный файл исчез из S3
    - Загружает новые документы
    - Возвращает: были ли изменения
    """
    s3 = boto3.client("s3", **s3_config)

    # 1. Получаем список файлов из S3
    response = s3.list_objects_v2(Bucket=bucket, Prefix=prefix)
    s3_keys = set()
    for obj in response.get("Contents", []):
        key = obj["Key"]
        if not key.endswith("/"):
            s3_keys.add(key)

    # 2. Получаем список уникальных source из Mongo
    mongo_sources = set(collection.distinct("metadata.source"))

    # 3. Удаляем из Mongo документы, которых больше нет на S3
    to_delete = mongo_sources - s3_keys
    # print(f'To delete: {to_delete}')
    logging.debug(f'To delete: {to_delete}')
    updated = False

    for missing_file in to_delete:
        logging.debug(f"[CLEANUP] Удаляю чанки из Mongo для: {missing_file}")
        # print(f"[CLEANUP] Удаляю чанки из Mongo для: {missing_file}")
        collection.delete_many({"metadata.source": missing_file})
        updated = True

    # 4. Обрабатываем новые файлы (лениво, по одному)
    with TemporaryDirectory() as temp_dir:
        for key in s3_keys:
            if key in mongo_sources:
                logging.debug(f"[SKIP] Уже обработано: {key}")
                # print(f"[SKIP] Уже обработано: {key}")
                continue
            
            logging.info(f"[PROCESS] Новый файл: {key}")
            # print(f"[PROCESS] Новый файл: {key}")
            local_path = os.path.join(temp_dir, os.path.basename(key))
            s3.download_file(bucket, key, local_path)

            if key.lower().endswith(".pdf"):
                loader = PyPDFLoader(local_path)
            elif key.lower().endswith(".txt"):
                loader = TextLoader(local_path)
            else:
                logging.warning(f"[SKIP] Неизвестный тип файла: {key}")
                # print(f"[SKIP] Неизвестный тип файла: {key}")
                continue

            docs = []
            for doc in loader.load():
                doc.page_content = clean_text(doc.page_content)
                doc.metadata["source"] = key
                docs.append(doc)

            chunks = split_documents(docs)
            compute_and_store_embeddings(chunks)
            updated = True

    return updated


def split_documents(documents):
    splitter = RecursiveCharacterTextSplitter(chunk_size=CHUNK_SIZE, chunk_overlap=CHUNK_OVERLAP)
    return splitter.split_documents(documents)


def compute_and_store_embeddings(chunks: List[Document]):
    embeddings = []
    new_chunks = []
    for chunk in tqdm(chunks, desc="Embedding"):
        text = chunk.page_content
        metadata = chunk.metadata
        if collection.find_one({"text": text}):
            continue  # уже есть
        try:
            vector = embedding_model.embed_query(text)
            doc = {
                "text": text,
                "embedding": vector,
                "metadata": metadata,
            }
            collection.insert_one(doc)
            embeddings.append(vector)
            new_chunks.append(chunk)
        except:
            return None
    return embeddings, new_chunks


def build_faiss_from_mongo():
    logging.info("[FAISS] Чтение эмбеддингов из MongoDB...")
    # print("[INFO] Чтение эмбеддингов из MongoDB...")
    embeddings = []
    documents = []

    for i, doc in enumerate(collection.find({}, {"embedding": 1, "text": 1, "metadata": 1})):
        vector = doc["embedding"]
        text = doc["text"]
        metadata = doc.get("metadata", {})
        document = Document(page_content=text, metadata=metadata)
        documents.append(document)
        embeddings.append(vector)

    logging.info("[FAISS] Построение FAISS индекса ...")
    # print("[INFO] Построение FAISS вручную...")
    dim = len(embeddings[0])
    index = faiss.IndexFlatL2(dim)
    index.add(np.array(embeddings).astype("float32"))

    docstore = InMemoryDocstore({str(i): documents[i] for i in range(len(documents))})
    index_to_docstore_id = {i: str(i) for i in range(len(documents))}

    return FAISS(
        embedding_function=embedding_model,
        index=index,
        docstore=docstore,
        index_to_docstore_id=index_to_docstore_id
    )


def init_vectorstore():
    global vectorstore
    logging.info("[INFO] Синхронизация Mongo <-> S3")
    # print("[INFO] Синхронизация Mongo <-> S3")
    updated = sync_mongo_with_s3(S3_BUCKET, S3_PREFIX, S3_CONFIG)

    if INDEX_DIR.exists() and not updated:
        logging.info("[INFO] Загружаю FAISS из диска (индекс не изменился)...")
        # print("[INFO] Загружаю FAISS из диска (индекс не изменился)...")
        vectorstore = FAISS.load_local(str(INDEX_DIR), embeddings=embedding_model, allow_dangerous_deserialization=True)
    else:
        logging.info("[INFO] Перестраиваю FAISS из Mongo...")
        # print("[INFO] Перестраиваю FAISS из Mongo...")
        vectorstore = build_faiss_from_mongo()
        vectorstore.save_local(str(INDEX_DIR))

    logging.info("[INFO] Индекс готов.")
    # print("[INFO] Индекс готов.")



def build_prompt(query: str, docs: List[Document]) -> str:
    parts = []
    for i, doc in enumerate(docs):
        # print(doc.metadata.keys())  # ← покажет точные ключи
        # print(doc.metadata)         # ← покажет всю структуру
        source = doc.metadata.get("source", "неизвестный документ")
        parts.append(f"[{i+1}] Источник: {source}\n{doc.page_content}")
    context = "\n\n".join(parts)
    return f"""Используй следующий контекст для ответа на вопрос.

Контекст:
{context}

Вопрос:
{query}

Ответ:"""


# --- Модель запроса ---
class QueryRequest(BaseModel):
    query: str
    k: int = 3


class NotifyRequest(BaseModel):
    key: str  # S3 путь к файлу


def extraction_data(docs):

    result = {
        'list_docs': []
    }
    for doc in docs:
        source = doc.metadata.get("source", "invalid")
        if source == "invalid":
            continue

        text = doc.page_content
        result['list_docs'].append({
            'doc_name': source,
            'text_fragment': text
        })
    return result


# --- Endpoint ---
@app.post("/query")
def query_rag(request: QueryRequest):
    with vectorstore_lock:
        vs = vectorstore

    docs = vs.similarity_search(request.query, k=request.k)
    result = extraction_data(docs)
    # prompt = build_prompt(request.query, docs)
    return result


@app.post("/delete")
def delete_file_from_db(request: NotifyRequest):
    # to_delete = request.key
    # # for missing_file in to_delete:
    # #     print(f"[CLEANUP] Удаляю чанки из Mongo для: {missing_file}")
    try:
        collection.delete_many({"metadata.source": request.key})
        logging.info(f"[DELETED]: {request.key}")
        # print(f"[DELETED]: {request.key}")
    except HTTPException as e:
        logging.error(e)
        # print(e)
        raise HTTPException(status_code=500, detail="Can't delete file")
        # updated = True


@app.post("/add")
def notify_new_document(request: NotifyRequest):
    key = request.key

    # Проверяем, есть ли уже такой файл в Mongo
    if collection.count_documents({"metadata.source": key}) > 0:
        return {"status": "skipped", "message": f"{key} уже обработан."}

    logging.info(f"[NOTIFY] Обработка нового документа: {key}")
    # print(f"[NOTIFY] Обработка нового документа: {key}")

    # Скачиваем файл
    with TemporaryDirectory() as temp_dir:
        local_path = os.path.join(temp_dir, os.path.basename(key))

        s3 = boto3.client("s3", **S3_CONFIG)
        try:
            s3.download_file(S3_BUCKET, key, local_path)
        except Exception as e:
            raise HTTPException(status_code=404, detail=f"Ошибка при скачивании {key}: {str(e)}")

        # Выбираем загрузчик
        if key.lower().endswith(".pdf"):
            loader = PyPDFLoader(local_path)
        elif key.lower().endswith(".txt"):
            loader = TextLoader(local_path)
        else:
            raise HTTPException(status_code=400, detail=f"Неподдерживаемый формат: {key}")

        # Обработка
        docs = []
        for doc in loader.load():
            doc.page_content = clean_text(doc.page_content)
            doc.metadata["source"] = key
            docs.append(doc)

        chunks = split_documents(docs)
        compute_and_store_embeddings(chunks)

    # # Перестроим FAISS (можно — только при необходимости)
    # global vectorstore
    # vectorstore = build_faiss_from_mongo()
    # vectorstore.save_local(str(INDEX_DIR))

    return {"status": "success", "message": f"{key} обработан и добавлен в базу"}







vectorstore_lock = Lock()

def start_faiss_rebuilder():
    # print("[FAISS] 🔁 Поток запущен")
    # time.sleep(FAISS_REBUILD_INTERVAL)

    def background_rebuild():
        logging.info("[FAISS] 🔁 Поток запущен")
        # print("[FAISS] 🔁 Поток запущен")
        time.sleep(FAISS_REBUILD_INTERVAL)
        global vectorstore
        while True:
            logging.info("[FAISS] ⏳ Плановая пересборка индекса из Mongo...")
            # print("[FAISS] ⏳ Плановая пересборка индекса из Mongo...")
            new_vs = build_faiss_from_mongo()
            new_vs.save_local(str(INDEX_DIR))

            with vectorstore_lock:
                vectorstore = new_vs  # безопасная подмена

            logging.info("[FAISS] ✅ Индекс обновлён")
            # print("[FAISS] ✅ Индекс обновлён")
            time.sleep(FAISS_REBUILD_INTERVAL)

            # time.sleep(FAISS_REBUILD_INTERVAL)

    # Запускаем в фоне как daemon
    thread = threading.Thread(target=background_rebuild, daemon=True)
    thread.start()




