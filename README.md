XRUST_BEZE backend repo
By: Petr09Mitin, endloc

Run:
- make dev-init (ONLY first time)
- make start


### ML Explanator API:  
ENDPOINT: http://ml_explanator:8091/explane  
Тип: POST  
```json
{  
  "query": "Запрос пользователя (может быть пустым)", 
  "answer": "Ответ эксперта."
}  
```
Возвращает:
```json
{ 
  "explanation": "текст в формате Markdown"
}
```

Ошибки сервера:  
* 400 - Answer cannot be empty. (Ответ эксперта - обязательное поле)
* 500 - Can't explain by api. (Ошибка во время обращения к API)
* 500 - Json decode error. (Ошибка во время парсинга ответа нейросети, в этом случае может помочь повторный запрос)
