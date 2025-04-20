from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import fitz


import mistral_api
import s3_utils

logging.basicConfig(level=logging.INFO)


app = FastAPI()


class TextInput(BaseModel):
    file_id: str
    backet_name: str

def extract_text_from_pdf(pdf_path):
    text = ""
    with fitz.open(pdf_path) as doc:
        for page in doc:
            text += page.get_text()
    return text



@app.post("/set-tag")
async def tag(input_data: TextInput):
    if not input_data.file_id.strip() or not input_data.backet_name.strip():
        raise HTTPException(status_code=400, detail="Fields can't be empty")

    local_path = f"/tmp/{input_data.file_id}"

    success = s3_utils.download_file_from_s3(input_data.backet_name, input_data.file_id, local_path)

    if not success:
        logging.error('Something went wrong')
        raise HTTPException(status_code=500, detail="Can't download file from S3")

    exstracted_text = extract_text_from_pdf(local_path)

    logging.info(f'text[:1000]: {exstracted_text[:1000]}')

    if len(exstracted_text) == 0:
        raise HTTPException(status_code=500, detail="Can't extract text from .pdf")


    tag = await mistral_api.set_tag(exstracted_text.strip())

    return tag