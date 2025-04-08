import streamlit as st
import requests
import os


# EXPLANATION_API_URL = os.environ.get('EXPLANATION_API_URL')
EXPLANATION_API_URL = "http://localhost:8066/check"
st.title("Содержит ли текст оскорбление")

text_input = st.text_area("Введите текст:")


if st.button("Check"):
    if text_input.strip():
        with st.spinner("Обработка запроса..."):
            try:
                response = requests.post(EXPLANATION_API_URL, json={"text": text_input})
                if response.status_code == 200:
                    st.success("Расширенный ответ:")
                    st.write(response.json().get("is_profanity", "Нет данных"))
                    st.write(response.json().get("swearing_list"))
                    st.dataframe(response.json().get("result"))
                    st.markdown("### Как выглядит json:")
                    st.write(response.json())
                else:
                    st.error(f"Ошибка: {response.status_code} - {response.json().get('detail')}")
            except Exception as e:
                st.error(f"Произошла ошибка при отправке запроса: {e}")
    else:
        st.warning("Пожалуйста, введите текст перед отправкой.")