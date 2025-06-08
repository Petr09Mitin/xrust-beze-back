from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import json


import mistral_api
import search

logging.basicConfig(level=logging.INFO)


app = FastAPI()


class TextInput(BaseModel):
    query: str
    answer: str


@app.post("/explane")
async def summarize(input_data: TextInput):
    if not input_data.answer.strip():
        raise HTTPException(status_code=400, detail="Answer cannot be empty.")

    try:
        explanation = await mistral_api.summarize(input_data.query.strip(), input_data.answer.strip())
    except Exception as e:
        raise HTTPException(status_code=500, detail="Can't explain by api.")

    # logging.info(f'Explanation of {explanation}.')

    try:
        # Данные вернулись в формате json
        data = json.loads(explanation)
    except json.decoder.JSONDecodeError:
        raise HTTPException(status_code=500, detail="Json decode error.")

    explanation_text = data["addition"]
    search_keywords = data["keywords"]

    # Если ключевых слов много - ограничимся 7ю
    search_keywords_main = search_keywords if len(search_keywords) <= 7 else search_keywords[:7]
    search_keywords_string = " ".join(search_keywords_main)

    array_of_names_links = await search.search_google(search_keywords_string)

    # Красиво отформатируем ссылки в Markdown
    mk_links = "\n\n## Полезные ссылки  \n"
    for i, link in enumerate(array_of_names_links):

        mk_links += f'{i+1}. [{link[0]}]({link[1]} Нажмите, чтобы перейти на сайт)  \n'
    
    

    # Объединим описание и ссылки
    all_text = explanation_text + '\n  ' + mk_links
    return {"explanation": all_text}
