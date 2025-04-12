import opennsfw2 as n2


def check_image(im_path: str) -> (bool, float):
    # Сохраняем временный файл

    prob = n2.predict_images([im_path])[0]
    is_nsfw = prob > 0.7

    return is_nsfw, prob
