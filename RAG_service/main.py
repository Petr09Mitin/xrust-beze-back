import os
import json
import faiss
import boto3
import fitz  # PyMuPDF
from docx import Document
from fastapi import FastAPI, Query
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

# --- Настройки ---
MODEL_NAME = "all-MiniLM-L6-v2"
INDEX_PATH = "index.faiss"
META_PATH = "doc_metadata.json"
BUCKET_NAME = "raw-data"
LOCAL_TMP_DIR = "tmp"

# --- Инициализация ---
os.makedirs(LOCAL_TMP_DIR, exist_ok=True)
s3 = boto3.client("s3",
    aws_access_key_id=os.getenv("MINIO_ROOT_USER"),
    aws_secret_access_key=os.getenv("MINIO_ROOT_PASSWORD"),
    endpoint_url=os.getenv("S3_ENDPOINT_URL")  # можно использовать кастомный URL для S3-совместимого хранилища
)
model = SentenceTransformer(MODEL_NAME)

app = FastAPI()

# --- Глобальные переменные ---
index = None
metadata = None

# --- Работа с индексом ---
def load_index():
    if os.path.exists(INDEX_PATH):
        return faiss.read_index(INDEX_PATH)
    return faiss.IndexFlatL2(model.get_sentence_embedding_dimension())

def save_index(idx):
    faiss.write_index(idx, INDEX_PATH)

def load_metadata():
    if os.path.exists(META_PATH):
        with open(META_PATH, "r", encoding="utf-8") as f:
            return json.load(f)
    return {}

def save_metadata(meta):
    with open(META_PATH, "w", encoding="utf-8") as f:
        json.dump(meta, f, ensure_ascii=False, indent=2)

# --- Работа с файлами ---
def list_s3_files(prefix=""):
    response = s3.list_objects_v2(Bucket=BUCKET_NAME, Prefix=prefix)
    return [obj["Key"] for obj in response.get("Contents", []) if obj["Key"].endswith((".pdf", ".docx"))]

def download_file(key, local_path):
    s3.download_file(BUCKET_NAME, key, local_path)

def extract_text(path):
    if path.endswith(".pdf"):
        return "\n".join([p.get_text() for p in fitz.open(path)])
    elif path.endswith(".docx"):
        return "\n".join([p.text for p in Document(path).paragraphs])
    return ""

def chunk_text(text, size=1500):
    return [text[i:i+size] for i in range(0, len(text), size)]

# --- Индексация ---
def sync_index_from_s3():
    global index, metadata
    new_metadata = {}

    for key in list_s3_files():
        if key in metadata:
            continue

        local_path = os.path.join(LOCAL_TMP_DIR, os.path.basename(key))
        download_file(key, local_path)

        text = extract_text(local_path)
        chunks = chunk_text(text)
        if not chunks:
            continue

        embeddings = model.encode(chunks, convert_to_numpy=True)
        index.add(embeddings)
        new_metadata[key] = {"chunks": len(chunks)}

        os.remove(local_path)

    metadata.update(new_metadata)
    save_metadata(metadata)
    save_index(index)
    return len(new_metadata)

# --- Поиск ---
def search_local_docs(query: str, top_k: int = 5):
    global index, metadata
    query_vec = model.encode([query])
    D, I = index.search(query_vec, top_k)

    flat_index = 0
    results = []
    for key, meta in metadata.items():
        for i in range(meta["chunks"]):
            if flat_index in I[0]:
                results.append(key)
            flat_index += 1
    return list(set(results))

# --- API ---
class QueryRequest(BaseModel):
    query: str
    top_k: int = 5

@app.post("/search")
def query_documents(request: QueryRequest):
    matches = search_local_docs(request.query, request.top_k)
    return {"matches": matches}

@app.post("/sync")
def sync_documents():
    new_docs = sync_index_from_s3()
    return {"new_documents_added": new_docs}

@app.on_event("startup")
def on_startup():
    global index, metadata
    index = load_index()
    metadata = load_metadata()

@app.get("/")
def root():
    return {"status": "RAG system with S3 is ready."}
