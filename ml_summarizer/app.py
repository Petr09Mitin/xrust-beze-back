from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

import mistral_api

app = FastAPI()


class TextInput(BaseModel):
    text: str


async def summarize_text(text: str) -> str:
    summarized_text = await mistral_api.summarize(text)
    # Здесь будет асинхронная логика для обработки текста
    # await asyncio.sleep(2)  # Имитация задержки для демонстрации асинхронности
    return f"Summary: {summarized_text}"  # Простая имитация суммаризации


@app.post("/summarize")
async def summarize(input_data: TextInput):
    if not input_data.text.strip():
        raise HTTPException(status_code=400, detail="Text field cannot be empty.")

    summary = await summarize_text(input_data.text)
    return {"summary": summary}
