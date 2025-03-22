import streamlit as st
import requests
import os


EXPLANATION_API_URL = os.environ.get('EXPLANATION_API_URL')

st.title("AI дополнение ответа")

query_input = st.text_area("Введите вопрос:")

answer_input = st.text_area("Введите ответ эксперта:", height=200)

if st.button("Дополнить ответ"):
    if answer_input.strip():
        with st.spinner("Обработка запроса..."):
            try:
                response = requests.post(EXPLANATION_API_URL, json={"query": query_input, "answer": answer_input})
                if response.status_code == 200:
                    st.success("Расширенный ответ:")
                    st.write(response.json().get("explanation", "Нет данных"))
                else:
                    st.error(f"Ошибка: {response.status_code} - {response.json().get('detail')}")
            except Exception as e:
                st.error(f"Произошла ошибка при отправке запроса: {e}")
    else:
        st.warning("Пожалуйста, введите текст перед отправкой.")