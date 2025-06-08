from mistralai import Mistral
import os
import logging
import re


import prompts


logging.basicConfig(level=logging.INFO)

MODEL_API_KEY = os.environ.get('MODEL_API_KEY')

client = Mistral(api_key=MODEL_API_KEY)


async def summarize(query, answer, RAG_context="", model="mistral-small-latest"):
    prompt = ''
    if query:
        prompt += f'Исходный запрос пользователя: "{query}".\n'
    
    prompt += f'Ответ эксперта: "{answer}".\n'
        # prompt = (f'Исходный запрос пользователя: "{query}".\n'
        #           f'Ответ эксперта: "{answer}".')
    # else:
    #     prompt = answer
    if len(RAG_context) > 0:
        prompt += f'\nФрагменты релевантных документов: {RAG_context}.'

    messages = [{"role": "system",
                 "content": prompts.system_prompt},
                {"role": "user", "content": prompt}]
    chat_response = client.chat.complete(model=model, messages=messages, )
    response_text = chat_response.choices[0].message.content
    response_json = re.search(r'\{.*\}', response_text, re.DOTALL).group()
    return response_json
