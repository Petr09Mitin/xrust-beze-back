from mistralai import Mistral
import os
import asyncio


API_KEY = os.environ.get('API_KEY')

# API_KEY = 'DbSIW6cjWpDFKpH2QDuZVUrcRRC58MyB'
client = Mistral(api_key=f'{API_KEY}')  # Тут вставляете свой ключ


async def summarize(user_message, model="mistral-small-latest"):
    messages = [{"role": "user", "content": user_message}]
    # print(client.models.list())
    chat_response = client.chat.complete( model=model, messages=messages, )
    return chat_response.choices[0].message.content


# print(run_mistral("Отвечай на русском языке. Сколько звезд на небе ?"))