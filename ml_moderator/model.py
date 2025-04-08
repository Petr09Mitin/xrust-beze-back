import numpy as np
import pandas as pd


class TextModerator:
    def __init__(self, model, vectorizer, text=None):
        self.model = model
        self.vectorizer = vectorizer
        self.text = text
        self.splited_text = []

    def add_text(self, text):
        self.text = text
        self.splited_text = []

    def split_text(self, num_words=5):
        words = self.text.split()
        self.splited_text = [words[i: min(i + 5, len(words))] for i in range(0, len(words), 5)]
        self.splited_text = [" ".join(x) for x in self.splited_text]

    def predict(self, text, threshold=0.7, num_words=5, overlap=1):
        stride = num_words - overlap
        assert stride > 0

        words = text.split()
        splited_text = [words[i: min(i + num_words, len(words))] for i in range(0, len(words), stride)]
        splited_text = [" ".join(x) for x in splited_text]
        tokenized_text = self.vectorizer.transform(splited_text)
        predicts_proba = self.model.predict_proba(tokenized_text)
        # print(predicts_proba.shape)

        positive_probs = predicts_proba[:, 1]
        true_idxs = np.where(positive_probs >= threshold)
        splited_text = np.array(splited_text)
        true_chanks = splited_text[true_idxs]
        result = pd.DataFrame({'text': splited_text, 'is_normal_prob': predicts_proba[:, 0],'is_swear_prob': positive_probs})

        return true_chanks, result

