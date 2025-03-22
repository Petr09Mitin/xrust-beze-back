FROM python:3.12-slim

# Устанавливаем рабочую директорию
WORKDIR /app

# Скопируем файлы проекта в контейнер
COPY . /app

# Устанавливаем зависимости
RUN pip install --no-cache-dir -r requirements.txt

# Команда для запуска FastAPI сервера
CMD ["streamlit", "run", "app.py"]