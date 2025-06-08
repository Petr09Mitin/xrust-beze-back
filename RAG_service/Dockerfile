FROM python:3.12-slim

# Устанавливаем рабочую директорию
WORKDIR /app

COPY ./RAG_service/requirements.txt /app

RUN pip install --no-cache-dir -r requirements.txt

# Скопируем файлы проекта в контейнер
COPY ./RAG_service /app

# Команда для запуска FastAPI сервера
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8929"]