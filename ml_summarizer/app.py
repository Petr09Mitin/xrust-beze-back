from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging

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
    summary_text, search_keywords = mistral_api.extract_text(summary)
    logging.info(summary_text)
    logging.info(search_keywords)
    links = search.search_google(search_keywords)
    # logging.info(f"Неочищенные: {links}.")
    # logging.info(f"Очищенные: {mistral_api.clean_text(search_keywords)}.")
    # logging.info(mistral_api.clean_text(search_keywords))
    # search_keywords = await mistral_api.get_links(summary)
    # links = search.search_google(search_keywords)
    string_links = '\n  \n'.join(links)
    logging.info(string_links)
    all_text = summary + '\n  ' + string_links
    # logging.info(search_keywords)
    # logging.info(string_links)
    return {"summary": all_text}
