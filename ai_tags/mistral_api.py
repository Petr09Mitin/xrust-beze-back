from mistralai import Mistral
from fastapi import FastAPI, HTTPException
import os
import logging

import json
import collections
import re


import prompts


logging.basicConfig(level=logging.INFO)

MODEL_API_KEY = os.environ.get('MODEL_API_KEY')

client = Mistral(api_key=MODEL_API_KEY)

valid_tags = set(['analytics', 'backend', 'architecture', 'database', 'design', 'devops', 'hardware', 'frontend', 'gamedev', 'integration', 'natural_languages', 'management', 'tools_for_business', 'ml', 'mobile', 'tools', 'testing', 'lowcode', 'math', 'security', 'other'])


# Загружаем дерево категорий
with open('skills_by_category.json', 'r', encoding='utf-8') as f:
    json_data = f.read()

categories = json.loads(json_data)


import re
from typing import List, Dict

def extract_relevant_skills(
    text: str,
    parent_category: str,
    *,
    category_to_skills: List[Dict] = categories,
    min_occurrences: int = 2,
    tf_threshold: float = 0.01,
    min_skill_length: int = 1
) -> List[str]:
    """
    Возвращает список навыков из указанной категории, которые встречаются в тексте
    и удовлетворяют условиям по частоте и количеству вхождений.
    """
    
    skills = []
    for d in category_to_skills:
        if d['category'] == parent_category:
            skills = d['skills']
            break
    
    logging.info(f'skill_list: {skills}')

    total_words = len(re.findall(r'\w+', text))
    relevant_skills = []

    for skill in skills:
        if len(skill) < min_skill_length:
            continue
        count = len(re.findall(rf'(?<!\w){re.escape(skill)}(?!\w)', text, flags=re.IGNORECASE))
        if count >= min_occurrences and (count / total_words) <= tf_threshold:
            relevant_skills.append(skill)

    return relevant_skills


async def set_tag(text, model="mistral-small-latest", max_len=1000):
    if len(text) == 0:
        raise HTTPException(status_code=500, detail="Text can't be empty")
    
    else:

        # postprompt = generate_category_hint(text)

        if len(text) > max_len:
            text = text[:max_len]
        
        messages = [{"role": "system",
                 "content": prompts.system_prompt},
                {"role": "user", "content": text}]
    
        logging.info(f'messages: {messages}')

        try:        
            chat_response = client.chat.complete(model=model, messages=messages, )
            response_text = chat_response.choices[0].message.content
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to get model response: {e}")

        # response_text = ""

        logging.info(f'API response: {response_text}')

        if not response_text:
            raise HTTPException(status_code=500, detail="Model response is empty")


        list_response = response_text.strip().split('\n')

        logging.info(f'list_response: {list_response}')

        is_study_material = list_response[0]

        if is_study_material == "yes":
            is_study_material_bool = True
            tag = list_response[1]
            name = list_response[2]
        elif is_study_material == "no":
            is_study_material_bool = False
            tag = ""
            name = ""
        else:
            raise HTTPException(status_code=500, detail="Response 'is_study_material' is invalid")

        if is_study_material_bool and tag not in valid_tags:
            raise HTTPException(status_code=500, detail="Response tag is invalid")
        else:
            return is_study_material_bool, tag, name