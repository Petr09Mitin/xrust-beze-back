import random
import csv

# Пути к входным файлам
BAD_WORDS_FILE = 'bad_words.txt'
NORMAL_WORDS_FILE = 'normal_words.txt'
OUTPUT_DATASET_FILE = 'dataset_random_2.csv'


# Функция серьёзного искажения бранного слова
def distort_word(word, num_variants=5):
    replace_map = {'у': 'y', 'и': '1', 'о': '0', 'а': '@', 'е': '3', 'х': 'x', 'с': 'c', 'з': '3', 'a': '@', 'e': '3',
                   'i': '1', 'o': '0', 's': '$', 't': '7', 'b': '8'}
    variants = set()

    for _ in range(num_variants):
        w = list(word)
        # print(w)

        # 1. Пропуск случайных букв
        if len(w) > 3:
            indices_to_remove = random.sample(range(len(w)), k=random.randint(1, min(2, len(w) - 2)))
            w = [c for i, c in enumerate(w) if i not in indices_to_remove]

        # 2. Перестановка букв
        if len(w) > 3 and random.random() < 0.5:
            i = random.randint(0, len(w) - 2)
            w[i], w[i + 1] = w[i + 1], w[i]

        # 3. Заменить символы на похожие
        for i in range(len(w)):
            if w[i] in replace_map and random.random() < 0.5:
                w[i] = replace_map[w[i]]

        # 4. Повтор символов
        if len(w) > 2 and random.random() < 0.3:
            j = random.randint(1, len(w) - 2)
            w.insert(j, w[j])

        # 5. Вставка символов
        if random.random() < 0.4:
            insert_char = random.choice(['*', '_', '-', '~'])
            k = random.randint(1, len(w) - 1)
            w.insert(k, insert_char)

        distorted = ''.join(w)
        variants.add(distorted)

    return list(variants)


# Создание положительного примера с искажённым матом
def create_offensive_sample(normal_words, bad_word):
    sample_len = random.randint(4, 10)
    selected = random.sample(normal_words, sample_len)
    insert_pos = random.randint(0, len(selected))
    sample = selected[:insert_pos] + [bad_word] + selected[insert_pos:]
    return " ".join(sample)


# Создание нейтрального примера
def create_neutral_sample(normal_words):
    sample_len = random.randint(4, 10)
    return " ".join(random.sample(normal_words, sample_len))


# Основная функция генерации датасета
def generate_dataset(bad_words_file, normal_words_file, output_file, samples_per_bad_word=10):
    with open(bad_words_file, 'r', encoding='utf-8') as f:
        bad_words = [line.strip() for line in f if line.strip()]
    with open(normal_words_file, 'r', encoding='utf-8') as f:
        normal_words = [line.strip() for line in f if line.strip()]

    dataset = []

    for i in range(3):
        for bad_word in bad_words:
            distorted_forms = distort_word(bad_word, num_variants=6)
            for form in distorted_forms:
                for _ in range(samples_per_bad_word):
                    text = create_offensive_sample(normal_words, form)
                    dataset.append((text, 1))

    for _ in range(len(dataset)):
        text = create_neutral_sample(normal_words)
        dataset.append((text, 0))

    random.shuffle(dataset)

    with open(output_file, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(['text', 'label'])
        for row in dataset:
            writer.writerow(row)

    return output_file


# Запуск
generate_dataset(BAD_WORDS_FILE, NORMAL_WORDS_FILE, OUTPUT_DATASET_FILE)
