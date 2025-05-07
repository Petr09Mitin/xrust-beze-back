from mistralai import Mistral
from fastapi import HTTPException
import os
import logging


logging.basicConfig(level=logging.INFO)

MODEL_API_KEY = os.environ.get('MODEL_API_KEY')

client = Mistral(api_key=MODEL_API_KEY)


async def get_responce_from_LLM(messages, model="mistral-small-latest"):
    """Отправляет запрос к LLM по API"""

    logging.info(f'messages: {messages}')

    try:        
        chat_response = client.chat.complete(model=model, messages=messages, )
        response_text = chat_response.choices[0].message.content
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get model response: {e}")

    logging.info(f'API response: {response_text}')

    if not response_text:
        raise HTTPException(status_code=500, detail="Model response is empty")

    return response_text
