import pickle

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging


logging.basicConfig(level=logging.DEBUG)

import swearing
import model as md

app = FastAPI()

MODEL_PATH = 'logreg_model.pkl'
VECTORIZER_PATH = 'vect.pkl'

with open(MODEL_PATH, 'rb') as f:
    model = pickle.load(f)

with open(VECTORIZER_PATH, 'rb') as f:
    vectorizer= pickle.load(f)

text_moderator = md.TextModerator(model=model, vectorizer=vectorizer)

class TextInput(BaseModel):
    text: str


@app.post("/check_swearing")
async def summarize(input_data: TextInput):
    if not input_data.text.strip():
        raise HTTPException(status_code=400, detail="Answer cannot be empty.")

    try:
        logging.info(f'Начинаем обработку')
        # is_profanity, preds = swearing.is_russian_profanity(input_data.text.strip())
        is_profanity = swearing.is_profanity_text(input_data.text.strip())
        # logging.info(f"preds: {preds}")
    except Exception as e:
        return HTTPException(status_code=500, detail="Can't check.")

        # logging.info(f'Explanation of {explanation}.')

    return {"is_profanity": is_profanity}


@app.post("/check")
async def check(input_data: TextInput):
    if not input_data.text.strip():
        raise HTTPException(status_code=400, detail="Answer cannot be empty.")

    try:
        logging.info(f'Начинаем обработку')
        # is_profanity, preds = swearing.is_russian_profanity(input_data.text.strip())
        swearing_list, result = text_moderator.predict(input_data.text.strip())
        logging.info(f'list: {swearing_list}')
        is_swearing = True if len(swearing_list) > 0 else False

        logging.info(f'is_swearing: {is_swearing}')
        # logging.info(f"preds: {preds}")
    except Exception as e:
        return HTTPException(status_code=500, detail="Can't check.")

        # logging.info(f'Explanation of {explanation}.')

    return {"is_profanity": is_swearing,
            "swearing_list": swearing_list.tolist(),
            "result": result}
