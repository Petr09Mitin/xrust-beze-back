import requests
import os


GOOGLE_API_KEY = os.environ.get('GOOGLE_API_KEY')
GOOGLE_CX = os.environ.get('GOOGLE_CX')


# Функция поиска через Google Custom Search API
async def search_google(query, api_key=GOOGLE_API_KEY, cx=GOOGLE_CX):
    search_url = f"https://www.googleapis.com/customsearch/v1?q={query}&key={api_key}&cx={cx}"

    response = requests.get(search_url)
    results = response.json().get("items", [])

    array_of_names_links = []
    for res in results:
        array_of_names_links.append((res['title'], res['link']))
        # names.append(res['title'])
        # links.append()

    return array_of_names_links
