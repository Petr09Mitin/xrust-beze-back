from mistralai import Mistral
import os
import logging
import re

logging.basicConfig(level=logging.INFO)

API_KEY = os.environ.get('API_KEY')

client = Mistral(api_key=API_KEY)


async def summarize(query, answer, model="mistral-small-latest"):
    # logging.info(f"Query: {query}.")
    # logging.info(f"Answer: {answer}.")
    prompt = (f'Твоя задача — дополнить ответ другого эксперта на заданный вопрос.) '
              ''
              'Входные данные: '
              f'1. Вопрос пользователя (если в наличии): {query} '
              f'2. Ответ другого эксперта: {answer} '
              'Твоя задача: '
              '1. Дополнить ответ, указав моменты, которые эксперт упустил. '
              '2. Привести примеры для лучшего понимания темы. '
              '3. Сформировать поисковые запросы для поиска дополнительных материалов по теме. (!!!Очень важно обрамить запросы в конструкцию  &=)(=&  !!! Цифры и порядковые номера не писать!)'
              ''
              'Формат ответа (следуй указанному формату строго): '
              '**Дополнение ответа эксперта:** '
              '---'
              '[Опиши недостающие моменты, дополни информацию, разверни сложные аспекты.] '
              ''
              '**Упущенные, но важные моменты:** '
              '---'
              '[Перечисли или опиши аспекты, которые эксперт не учел.] '
              ''
              '**Примеры:**'
              '---'
              '[Приведи конкретные примеры, демонстрирующие основные идеи.] '
              ''
              '**Поисковые запросы для поиска статей:**'
              '&=)Перечисли поисковые запросы (ключевые слова), которые помогут найти качественные материалы в формате научных статей, тематических журналов, репозиториев и доступных pdf-документов(=&')

    messages = [{"role": "system",
                 "content": "Ты эксперт в области программирования, машинного обучения, Data Science, системного дизайна и других областях IT. "},
                {"role": "user", "content": prompt}]
    chat_response = client.chat.complete(model=model, messages=messages, )
    return chat_response.choices[0].message.content


async def get_links(text, model="mistral-small-latest"):
    system_prompt = ("Ты эксперт по генерации поисковых запросов для нахождения обучающих материалов, книг, гайдов и статей по программированию, разработке, машинному обучению и другим техническим темам. "
                     "Пользователь предоставит описание того, что он хочет изучить. На основе этого:"
                     ""
                     "Определи основную тему и несколько альтернативных формулировок."
                     "Сгенерируй поисковые запросы для нахождения:"
                     "Учебников и книг (добавь 'tutorial', 'guide', 'step-by-step', 'book', 'PDF')."
                     "Практических примеров и решений (добавь 'example', 'solution', 'code sample')."
                     "Ресурсов по ключевым темам на сайтах site:habr.com, site:stackoverflow.com, site:github.com, site:medium.com, site:realpython.com."
                     "Укажи фильтры по формату файла filetype:pdf или filetype:html, если нужно.")
    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": text}
    ]
    chat_response = client.chat.complete(model=model, messages=messages, )
    return chat_response.choices[0].message.content


def extract_text(data):
    # Регулярное выражение для поиска текста внутри конструкции
    pattern = r'&=\)(.*?)\(=&'

    # Текст внутри конструкции
    inside_text = re.search(pattern, data)
    inside_text = inside_text.group(1) if inside_text else ''

    # Удаляем текст внутри конструкции из исходного текста, чтобы получить только внешний текст
    outside_text = re.sub(pattern, '', data)

    return inside_text, outside_text


def clean_text(text):
    # Удаляем все символы, которые не являются буквами, цифрами или точками
    cleaned_text = re.sub(r'[^a-zA-Zа-яА-Я0-9.]', '', text)
    return cleaned_text
