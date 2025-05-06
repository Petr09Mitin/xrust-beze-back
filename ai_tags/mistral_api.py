from mistralai import Mistral
from fastapi import HTTPException
import os
import logging

# import json
# import collections
# import re


# import prompts


logging.basicConfig(level=logging.INFO)

MODEL_API_KEY = os.environ.get('MODEL_API_KEY')

client = Mistral(api_key=MODEL_API_KEY)



# # Загружаем дерево категорий
# with open('skills_by_category.json', 'r', encoding='utf-8') as f:
#     json_data = f.read()

# categories = json.loads(json_data)


# import re
# from typing import List, Dict

# def extract_relevant_skills(
#     text: str,
#     parent_category: str,
#     *,
#     category_to_skills: List[Dict] = categories,
#     min_occurrences: int = 2,
#     tf_threshold: float = 0.01,
#     min_skill_length: int = 1
# ) -> List[str]:
#     """
#     Возвращает список навыков из указанной категории, которые встречаются в тексте
#     и удовлетворяют условиям по частоте и количеству вхождений.
#     """
    
#     skills = []
#     for d in category_to_skills:
#         if d['category'] == parent_category:
#             skills = d['skills']
#             break
    
#     logging.info(f'skill_list: {skills}')

#     total_words = len(re.findall(r'\w+', text))
#     relevant_skills = []

#     for skill in skills:
#         if len(skill) < min_skill_length:
#             continue
#         count = len(re.findall(rf'(?<!\w){re.escape(skill)}(?!\w)', text, flags=re.IGNORECASE))
#         if count >= min_occurrences and (count / total_words) <= tf_threshold:
#             relevant_skills.append(skill)

#     return relevant_skills

async def get_responce_from_LLM(messages, model="mistral-small-latest"):

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
        

