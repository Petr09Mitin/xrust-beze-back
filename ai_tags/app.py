from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging
import fitz
import os
from docx import Document

import tags_utilities
import s3_utils

logging.basicConfig(level=logging.INFO)


app = FastAPI()


class TextInput(BaseModel):
    file_id: str
    bucket_name: str


def extract_text_from_pdf(pdf_path):
    text = ""
    with fitz.open(pdf_path) as doc:
        for page in doc:
            text += page.get_text()
    return text


def extract_text_from_docx(docx_path):
    text = ""
    doc = Document(docx_path)
    for paragraph in doc.paragraphs:
        text += paragraph.text + "\n"
    return text


def extract_text_from_txt(txt_path):
    with open(txt_path, 'r', encoding='utf-8') as f:
        return f.read()


def extract_text(file_path):
    ext = os.path.splitext(file_path)[1].lower()

    if ext == '.pdf':
        return extract_text_from_pdf(file_path)
    elif ext == '.docx':
        return extract_text_from_docx(file_path)
    elif ext == '.txt':
        return extract_text_from_txt(file_path)
    else:
        raise ValueError(f"Unsupported file type: {ext}")


@app.post("/set-tag")
async def tag(input_data: TextInput):
    try:
        if not input_data.file_id.strip() or not input_data.bucket_name.strip():
            raise HTTPException(status_code=400, detail="Fields can't be empty")

        local_path = f"/tmp/{input_data.file_id}"

        success = s3_utils.download_file_from_s3(input_data.bucket_name, input_data.file_id, local_path)

        if not success:
            logging.error('Something went wrong')
            raise HTTPException(status_code=400, detail="Can't download file from S3")

        extracted_text = extract_text(local_path)

        logging.info(f'text[:1000]: {extracted_text[:1000]}')

        if len(extracted_text) == 0:
            raise HTTPException(status_code=400, detail="Can't extract text from .pdf")

        is_study_material_bool, tag_list, name = await tags_utilities.set_tags(extracted_text.strip())

        response = {"is_study_material": is_study_material_bool,
                    "study_material": {
                        "name": name,
                        "tags": tag_list
                    }}
        
        return response
    
    except Exception as e:
        logging.error(e)
        raise e
    
    finally:
        os.remove(local_path)
        if not os.path.exists(local_path):
            logging.info(f'Файл успешно удалён')
