from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import json

logging.basicConfig(level=logging.INFO)

import mistral_api
import search

app = FastAPI()


class TextInput(BaseModel):
    query: str
    answer: str


@app.post("/summarize")
async def summarize(input_data: TextInput):
    if not input_data.answer.strip():
        raise HTTPException(status_code=400, detail="Text field cannot be empty.")

    summary = await mistral_api.summarize(input_data.query.strip(), input_data.answer)
    logging.info(f'Summary for {summary}.')

    data = json.loads(summary)

    summary_text = data["addition"]
    search_keywords = " ".join(data["keywords"])
    search_keywords_main = search_keywords if len(search_keywords) <= 7 else search_keywords[:7]





    # summary_text, search_keywords = mistral_api.extract_text(summary)
    # logging.info(summary_text)
    # logging.info(search_keywords_main)
    links = search.search_google(search_keywords_main)
    logging.info(f"Неочищенные: {links}.")
    # logging.info(f"Очищенные: {mistral_api.clean_text(search_keywords)}.")
    # logging.info(mistral_api.clean_text(search_keywords))
    # search_keywords = await mistral_api.get_links(summary)
    # links = search.search_google(search_keywords)
    mk_links = "\n\n## Полезные ссылки:  \n"
    for i, link in enumerate(links):
        mk_links += f'{i+1}. {link}  \n'
    # string_links = '\n  \n'.join(links)
    # logging.info(string_links)
    all_text = summary_text + '\n  ' + mk_links
    # logging.info(search_keywords)
    # logging.info(string_links)
    return {"summary": all_text}
