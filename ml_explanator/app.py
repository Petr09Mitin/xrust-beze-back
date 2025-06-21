from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import json
import os
import requests


import mistral_api
import search

logging.basicConfig(level=logging.INFO)


app = FastAPI()


RAG_URL = os.getenv('RAG_URL')


class TextInput(BaseModel):
    query: str
    answer: str

def get_rag_answer(query, k=5):
    try:
        RAG_context = requests.post(RAG_URL, json={
                                                    "query": query,
                                                    "k": k
                                                })
        if RAG_context.status_code != 200:
            return "", []
        
        # logging.info(f'[RAG_ANSWER]: {RAG_context.json()}')

        list_docs = RAG_context.json()['list_docs']

        # logging.info(f'list_docs: {list_docs}')

        filenames = []
        fragments = []
        for doc in list_docs:
            # logging.info(f'type: {type(doc)}')
            # logging.info(f'doc: {doc}')
            docname = doc['doc_name']
            text_fragment = doc['text_fragment']

            # logging.info(f'docname: {docname}; text_fragment: {text_fragment}')

            if docname not in filenames:
                filenames.append(docname)
            
            fragments.append(text_fragment)
        
        fragments_formated = [f'[{i}] {fr}' for i, fr in enumerate(fragments, start=1)]
        text = '\n'.join(fragments_formated)
        # logging.info(f'text: {text}; filenames: {filenames}')
    except:
        return "", []
    return text, filenames



@app.post("/explane")
async def summarize(input_data: TextInput):
    if not input_data.answer.strip():   
        raise HTTPException(status_code=400, detail="Answer cannot be empty.")

    rag_context, filenames = get_rag_answer(input_data.answer.strip())

    # logging.info(f'[RAG_CONTEXT]: {rag_context}; {filenames}')

    # return {"explanation": rag_context,
    #         "relevant_files": filenames}

    try:
        explanation = await mistral_api.summarize(input_data.query.strip(), input_data.answer.strip(), rag_context)
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

        mk_links += f'{i+1}. [{link[0]}]({link[1]}) \n'
    

    mk_files = "\n\n## Возможно релевантные файлы из нашей базы знаний \n"

    for i, filename in enumerate(filenames, start=1):
        mk_files += f'- [Файл {i}.](https://skill-sharing.ru/materials/{filename}) \n'


    # Объединим описание и ссылки
    all_text = explanation_text + '\n  ' + mk_links
    if len(filenames) > 0:
        all_text += mk_files

    return {"explanation": all_text,
            "relevant_files": filenames}
