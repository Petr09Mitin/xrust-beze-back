FROM python:3.12-slim

# Устанавливаем рабочую директорию
WORKDIR /app

COPY ./ai_tags/requirements.txt /app

RUN pip install --no-cache-dir -r requirements.txt

# Скопируем файлы проекта в контейнер
COPY ./ai_tags /app

# Команда для запуска FastAPI сервера
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8087"]