import random
import csv

# Пути к файлам
BAD_WORDS_RU_FILE = 'bad_words.txt'
BAD_WORDS_EN_FILE = 'bad_words_en.txt'
NORMAL_WORDS_RU_FILE = 'normal_words.txt'
NORMAL_WORDS_EN_FILE = 'normal_words_en.txt'
OUTPUT_DATASET_FILE = 'mixed_dataset_2.csv'

# Мапа замен (расширена на англ)
REPLACE_MAP = {
    'у': 'y', 'и': '1', 'о': '0', 'а': '@', 'е': '3', 'х': 'x', 'с': 'c', 'з': '3',
    'a': '@', 'e': '3', 'i': '1', 'o': '0', 's': '$', 't': '7', 'b': '8'
}

# Искажение слова (поддерживает RU/EN)
def distort_word(word, num_variants=6):
    variants = set()
    for _ in range(num_variants):
        w = list(word)

        # Пропуск букв
        if len(w) > 3:
            remove_count = random.randint(1, min(2, len(w)-2))
            indices = random.sample(range(len(w)), remove_count)
            w = [c for i, c in enumerate(w) if i not in indices]

        # Перестановка
        if len(w) > 3 and random.random() < 0.5:
            i = random.randint(0, len(w)-2)
            w[i], w[i+1] = w[i+1], w[i]

        # Замена символов
        for i in range(len(w)):
            if w[i].lower() in REPLACE_MAP and random.random() < 0.5:
                w[i] = REPLACE_MAP[w[i].lower()]

        # Повтор буквы
        if len(w) > 2 and random.random() < 0.3:
            j = random.randint(1, len(w)-2)
            w.insert(j, w[j])

        # Вставка символа
        if random.random() < 0.4:
            w.insert(random.randint(1, len(w)-1), random.choice(['*', '_', '-', '~']))

        distorted = ''.join(w)
        variants.add(distorted)

    return list(variants)

# Смешанный контекст с бранным словом
def create_offensive_sample(normal_words, bad_word):
    sample_len = random.randint(4, 10)
    selected = random.sample(normal_words, sample_len)
    insert_pos = random.randint(0, len(selected))
    sample = selected[:insert_pos] + [bad_word] + selected[insert_pos:]
    return " ".join(sample)

# Нейтральный микс
def create_neutral_sample(normal_words):
    sample_len = random.randint(4, 10)
    return " ".join(random.sample(normal_words, sample_len))

# Генерация датасета
def generate_mixed_dataset(bad_ru, bad_en, norm_ru, norm_en, output_file, samples_per_bad_word=10):
    # Загрузка слов
    def load_file(path):
        with open(path, 'r', encoding='utf-8') as f:
            return [line.strip() for line in f if line.strip()]

    bad_words = load_file(bad_ru) + load_file(bad_en)
    normal_words = load_file(norm_ru) + load_file(norm_en)

    dataset = []

    for i in range(3):
        for bad_word in bad_words:
            distorted_forms = distort_word(bad_word, num_variants=6)
            for form in distorted_forms:
                for _ in range(samples_per_bad_word):
                    text = create_offensive_sample(normal_words, form)
                    dataset.append((text, 1))

    for _ in range(len(dataset)):
        dataset.append((create_neutral_sample(normal_words), 0))

    random.shuffle(dataset)

    with open(output_file, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(['text', 'label'])
        t=0
        for row in dataset:
            t+=1
            print(t)
            writer.writerow(row)

    print(f"✓ Датасет сохранён в: {output_file}")
    return output_file

# Запуск
generate_mixed_dataset(
    BAD_WORDS_RU_FILE,
    BAD_WORDS_EN_FILE,
    NORMAL_WORDS_RU_FILE,
    NORMAL_WORDS_EN_FILE,
    OUTPUT_DATASET_FILE
)
