FROM python:3.12-slim

# Устанавливаем зависимости системы
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только requirements.txt на этом этапе
COPY ./ml_image_moderator/requirements.txt /app

# Устанавливаем Python-зависимости
RUN pip install --no-cache-dir -r requirements.txt

# Теперь копируем весь остальной код
COPY ./ml_image_moderator /app

# Задаём команду по умолчанию
CMD ["python", "src/server.py"]
