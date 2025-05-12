from fastapi import UploadFile, FastAPI, HTTPException
from transformers import pipeline
from pydantic import BaseModel
import torch
import os
import uuid
import time
import logging

import s3_utils

class TextInput(BaseModel):
    file_id: str
    bucket_name: str

app = FastAPI()

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

# Настройка устройства
device = "cuda" if torch.cuda.is_available() else "cpu"
logging.info(f"Using device: {device}.")
print(f"Используемое устройство: {device}")

# Загружаем pipeline для Whisper
pipe = pipeline(
    "automatic-speech-recognition",
    model="openai/whisper-large-v3-turbo",
    device=device
)


# Функция для обработки аудио в фоне
def process_audio(audio_file: str):
    try:
        start_time = time.time()

        # 1. Транскрипция с помощью Whisper
        whisper_result = pipe(
            audio_file,
            return_timestamps=True,
            chunk_length_s=30,
            stride_length_s=5,
        )

        # 3. Подготовка результатов
        full_text = whisper_result["text"]

        timestamps = [f'[{chunk["timestamp"][0]}-{chunk["timestamp"][1]}] {chunk["text"]}' for chunk in whisper_result["chunks"]]
        timestamps_str = "\n".join(timestamps)


        end_time = time.time()
        print(f"Обработка файла {audio_file} заняла {end_time - start_time:.2f} секунд")

        return full_text, timestamps_str

    except Exception as e:
        logging.error(str(e))



# Эндпоинт для загрузки файла
@app.post("/transcribe/")
async def transcribe_audio(input_data: TextInput):
    try:
        if not input_data.file_id.strip() or not input_data.bucket_name.strip():
            raise HTTPException(status_code=400, detail="Fields can't be empty")
        
        local_path = f"/tmp/{input_data.file_id}"

        success = s3_utils.download_file_from_s3(input_data.bucket_name, input_data.file_id, local_path)

        if not success:
            logging.error('Something went wrong')
            raise HTTPException(status_code=400, detail="Can't download file from S3")
        
        text, text_ts = process_audio(local_path)
        logging.info(f'Extracted text: {text}')


        return {"text": text, "text_ts": text_ts}
    
    except Exception as e:
        logging.error(e)
        raise e
    
    finally:
        os.remove(local_path)
        if not os.path.exists(local_path):
            logging.info(f'Файл успешно удалён')
