import ahocorasick
import json
import re
from typing import List, Dict
import logging
import prompts
import mistral_api
from fastapi import FastAPI, HTTPException


valid_tags = set(['analytics', 'backend', 'architecture', 'database', 'design', 'devops', 'hardware', 'frontend', 'gamedev', 'integration', 'natural_languages', 'management', 'tools_for_business', 'ml', 'mobile', 'tools', 'testing', 'lowcode', 'math', 'security', 'other'])


# ------------Загружаем дерево категорий--------------
with open('skills_by_category.json', 'r', encoding='utf-8') as f:
    json_data = f.read()

categories = json.loads(json_data)
# ----------------------------------------------------

# --Строим поисковый движок на основе ключевых слов---
BMSTU_STRING = ['bmstu', 'мгту', 'баумана', 'иу1', 'иу2', 'иу3', 'иу4', 'иу5', 'иу6', 'иу7', 'иу8', 'иу9', 'иу10', 'иу11',
                'иу-1', 'иу-2', 'иу-3', 'иу-4', 'иу-5', 'иу-6', 'иу-7', 'иу-8', 'иу-9', 'иу-10', 'иу-11', 'информатика искусственный интеллект и системы управления', 'информатика и системы управления']

A = ahocorasick.Automaton()

for idx, key in enumerate(BMSTU_STRING):
    A.add_word(key, (idx, key))
A.make_automaton()
# ---------------------------------------------------

# ------Загружаем список Бауманских категорий--------
with open('bmstu_tags.txt', 'r') as f:
    bmstu_categories = [string.strip() for string in f.readlines()]
    bmstu_categories_string = f'[{', '.join(bmstu_categories)}]'
# ----------------------------------------------------

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


async def set_common_tag(text, max_len=1000):

    if len(text) > max_len:
        text = text[:max_len]
    
    messages = [{"role": "system", "content": prompts.common_system_prompt},
                {"role": "user", "content": text}
                ]

    response_text = await mistral_api.get_responce_from_LLM(messages=messages)

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
        raise HTTPException(status_code=500, detail="Response common _ag is invalid")
    else:
        return is_study_material_bool, tag, name


def is_bmstu_text(text):
    for end_index, (idx, found) in A.iter(text.lower()):
        return True
    return False


async def set_bmstu_tag(text, max_len=1000):
    if len(text) > max_len:
            text = text[:max_len]

    messages = [{"role": "user", "content": prompts.bmstu_system_prompt.format(disciplines=bmstu_categories_string, text=text)},
                ]
    
    response_text = await mistral_api.get_responce_from_LLM(messages=messages)

    list_response = response_text.strip().split('\n')

    if len(list_response) != 2 or list_response[0] not in bmstu_categories:
        raise HTTPException(status_code=500, detail="Response bmstu_tag is invalid")
    else:
        is_study_material_bool = True
        tag = list_response[0]
        name = list_response[1]

        return is_study_material_bool, tag, name


async def set_tags(text):
    if len(text) == 0:
        raise HTTPException(status_code=500, detail="Text can't be empty")
    
    max_len = 2000
    
    if is_bmstu_text(text):
        logging.info(f'Detected bmstu material')
        is_study_material_bool, tag, name = await set_bmstu_tag(text, max_len=max_len)
        bmstu_tag_name = 'МГТУ им. Н.Э. Баумана'
    else:
        logging.info(f'Common material')
        is_study_material_bool, tag, name = await set_common_tag(text, max_len=max_len)
        bmstu_tag_name = None
    
    logging.info(f'is_study_material_bool: {is_study_material_bool}, tag: {tag}, name: {name}.')

    if is_study_material_bool:
        additional_tags = extract_relevant_skills(text, tag)
        
        tag_list = []

        if bmstu_tag_name:
            tag_list += [bmstu_tag_name]

        tag_list += [tag]
        tag_list += additional_tags
    else:
        tag_list = []
        name = ""
    
    return is_study_material_bool, tag_list, name


# if __name__ == "__main__":
#     is_bmstu = is_bmstu_text(text)
#     print(is_bmstu)
    