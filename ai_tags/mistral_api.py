from mistralai import Mistral
import os
import logging
import re
import collections
import json


import prompts


logging.basicConfig(level=logging.INFO)

MODEL_API_KEY = os.environ.get('MODEL_API_KEY')

client = Mistral(api_key=MODEL_API_KEY)

valid_tags = set(['analytics', 'backend', 'architecture', 'database', 'design', 'devops', 'hardware', 'frontend', 'gamedev', 'integration', 'natural_languages', 'management', 'tools_for_buisness', 'ml', 'mobile', 'tools', 'testing', 'lowcode', 'math', 'security'])





with open('skill_to_category.json', 'r', encoding='utf-8') as f:
    json_data = f.read()

categories = json.loads(json_data)


def generate_category_hint(
    text: str, 
    skill_to_category: dict = categories, 
    min_occurrences: int = 2, 
    dominance_ratio: float = 0.5,
    tf_threshold: float = 0.01,
    min_skill_length: int = 1
) -> str:
    """
    Ищет в тексте навыки и формирует подсказку с фильтрацией общих слов:
    - Считает вхождения каждого навыка и общее число слов в тексте.
    - Фильтрует навыки длиной < min_skill_length или с относительной частотой > tf_threshold.
    - Среди оставшихся применяет логику min_occurrences и dominance_ratio.
    """
    # Подсчёт вхождений навыка
    skill_counts = {}
    for skill, category in skill_to_category.items():
        # Пропускаем короткие названия
        if len(skill) < min_skill_length:
            continue
        # Ищем с помощью lookaround, чтобы корректно работать с несловарными символами
        occurrences = len(re.findall(rf'(?<!\w){re.escape(skill)}(?!\w)', text, flags=re.IGNORECASE))
        if occurrences:
            skill_counts[skill] = occurrences
    
    if not skill_counts:
        return "No notable terms found in the text."
    
    # Общее количество "слов" в тексте
    total_words = len(re.findall(r'\w+', text))
    
    # Фильтрация по относительной частоте
    filtered = {s: cnt for s, cnt in skill_counts.items() if cnt / total_words <= tf_threshold}
    relevant = {s: cnt for s, cnt in (filtered or skill_counts).items() if cnt >= min_occurrences}
    if not relevant:
        relevant = filtered or skill_counts
    
    # Суммируем по категориям
    cat_counts = collections.Counter()
    for skill, cnt in relevant.items():
        cat_counts[skill_to_category[skill]] += cnt

    total_relevant = sum(cat_counts.values())
    top_cat, top_cnt = cat_counts.most_common(1)[0]
    
    # Формируем вывод
    terms_str = ', '.join(f"{skill}({cnt})" for skill, cnt in relevant.items())
    if top_cnt / total_relevant >= dominance_ratio:
        return (
            f"The text contains the following notable terms: {terms_str}. "
            f"These are usually associated with {top_cat}."
        )
    else:
        details = ', '.join(f"{s} ({skill_to_category[s]})" for s in relevant)
        cat_list = '; '.join(f"{cat}: {cnt}" for cat, cnt in cat_counts.items())
        return (
            f"The text contains the following notable terms: {details}. "
            f"These span multiple categories ({cat_list}). "
            "Please choose the single most appropriate category."
            "WARNING! This information is only auxiliary in controversial cases, make a decision based on the source text."
        )



async def set_tag(text, model="mistral-small-latest", max_len=1000):
    if len(text) == 0:
        raise "Text can't be empty"
    
    else:

        postprompt = generate_category_hint(text)

        if len(text) > max_len:
            text = text[:max_len]
        
        messages = [{"role": "system",
                 "content": prompts.system_prompt},
                {"role": "user", "content": text + "\n\n" + postprompt}]
    
        logging.info(f'messages: {messages}')
        
        chat_response = client.chat.complete(model=model, messages=messages, )
        response_text = chat_response.choices[0].message.content

        logging.info(f'API response: {response_text}')
        if not response_text:
            raise "Response is empty"
        if response_text not in valid_tags:
            raise "Response tag is invalid"
        else:
            return {"tag": response_text}