import random
import csv
import os

# Пути к входным файлам
BAD_WORDS_FILE = 'words.txt'
NORMAL_WORDS_FILE = 'normal_words.txt'
OUTPUT_DATASET_FILE = 'dataset.csv'

# Функция генерации искажений бранных слов
def distort_word(word):
    variants = {word}
    letters = list(word)
    replace_map = {'у':'y', 'и':'1', 'о':'0', 'а':'@', 'е':'3', 'х':'x', 'с':'c', 'з':'3'}

    if len(word) > 2:
        # Повтор случайной буквы
        pos = random.randint(1, len(word)-2)
        variants.add(word[:pos] + word[pos]*2 + word[pos+1:])
        # Заменить буквы на похожие
        for i, ch in enumerate(letters):
            if ch in replace_map:
                variant = word[:i] + replace_map[ch] + word[i+1:]
                variants.add(variant)
        # Вставка спецсимволов
        variants.add(word[0] + '*' + word[1:])
        variants.add(word[0] + '_' + word[1:])
        variants.add(word[0] + '-' + word[1:])
    return list(variants)

# Создание положительного примера с вкраплённым бранным словом
def create_offensive_sample(normal_words, bad_word):
    selected = random.sample(normal_words, 10)
    insert_pos = random.randint(0, len(selected))
    sample = selected[:insert_pos] + [bad_word] + selected[insert_pos:]
    return " ".join(sample)

# Создание нейтрального примера
def create_neutral_sample(normal_words):
    return " ".join(random.sample(normal_words, 11))

# Основная функция
def generate_dataset(bad_words_file, normal_words_file, output_file, samples_per_bad_word=10):
    # Чтение данных
    with open(bad_words_file, 'r', encoding='utf-8') as f:
        bad_words = [line.strip() for line in f if line.strip()]
    with open(normal_words_file, 'r', encoding='utf-8') as f:
        normal_words = [line.strip() for line in f if line.strip()]

    dataset = []

    for bad_word in bad_words:
        distorted_forms = distort_word(bad_word)
        for form in distorted_forms:
            for _ in range(samples_per_bad_word):
                text = create_offensive_sample(normal_words, form)
                dataset.append((text, 1))

    num_negative = len(dataset)
    for _ in range(num_negative):
        text = create_neutral_sample(normal_words)
        dataset.append((text, 0))

    random.shuffle(dataset)

    with open(output_file, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(['text', 'label'])
        for row in dataset:
            writer.writerow(row)

    return output_file

# Запускаем генерацию
generate_dataset(BAD_WORDS_FILE, NORMAL_WORDS_FILE, OUTPUT_DATASET_FILE)
