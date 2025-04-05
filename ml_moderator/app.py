from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging

logging.basicConfig(level=logging.DEBUG)

import swearing

app = FastAPI()


class TextInput(BaseModel):
    text: str


@app.post("/check")
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
