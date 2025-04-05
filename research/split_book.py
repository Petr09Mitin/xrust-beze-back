import re

def text_to_wordlist(input_path, output_path):
    with open(input_path, 'r', encoding='utf-8') as f:
        text = f.read()

    # Удаляем пунктуацию и разбиваем на слова
    words = re.findall(r'\b\w+\b', text.lower())

    # Сохраняем каждое слово на отдельной строке
    with open(output_path, 'w', encoding='utf-8') as f:
        for word in words:
            f.write(word + '\n')

# Пример использования
input_file = 'book_en.txt'           # замените на путь к исходному файлу
output_file = 'normal_words_en.txt'       # имя выходного файла
text_to_wordlist(input_file, output_file)
