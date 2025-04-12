from check_swear import SwearingCheck
import logging

sch = SwearingCheck()  # create filter


def is_russian_profanity(word: str):
    preds = sch.predict_proba(word)

    return preds[0]
    # if len(pred) == 1:
    #     if pred <= 0.3:
    #         return False
    #     else:
    #         return True


def slice_text_with_overlap(text, window_size=5, step=2):
    """
    Нарезает текст на фрагменты заданной длины (в словах) с перекрытием.

    Параметры:
    text (str): исходный текст
    window_size (int): количество слов в каждом фрагменте
    step (int): шаг сдвига окна (по умолчанию 1)

    Возвращает:
    list: список строк, где каждая строка состоит из window_size слов.
    """
    words = text.split()  # разбиваем текст на слова
    if len(words) < window_size:
        return [text]
    slices = []
    # Проходим по списку слов с шагом step и формируем фрагменты
    for i in range(0, len(words) - window_size + 1, step):
        window = words[i:i + window_size]
        slices.append(" ".join(window))
    return slices


def is_profanity_text(text: str, threshold=0.5):
    sentences = slice_text_with_overlap(text, window_size=5)
    logging.info(sentences)
    predicts = []
    for sentence in sentences:
        predicts.append(is_russian_profanity(sentence))

    logging.info(f"predicts: {predicts}.")
    if max(predicts) > threshold:
        return True
    return False


print(slice_text_with_overlap("NTFFJHF f"))