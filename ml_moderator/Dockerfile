FROM python:3.12-slim

# Устанавливаем рабочую директорию
WORKDIR /app

# Скопируем файлы проекта в контейнер
COPY .. /app

# Устанавливаем зависимости
# RUN pip install --no-cache-dir -r ml_moderator/requirements.txt // TODO !!!!!!
RUN pip install -r ml_moderator/requirements.txt

# Команда для запуска FastAPI сервера
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8066"]