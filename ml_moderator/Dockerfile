FROM python:3.12-slim

# Устанавливаем рабочую директорию
WORKDIR /app

# Скопируем файлы проекта в контейнер
COPY ./ml_moderator /app

# Устанавливаем зависимости
RUN pip install --no-cache-dir -r requirements.txt

# Команда для запуска FastAPI сервера
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8066"]