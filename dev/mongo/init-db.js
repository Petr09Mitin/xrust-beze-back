db = db.getSiblingDB("admin"); // Переключаемся на админскую БД
db.auth("admin", "admin")
xb_db = db.getSiblingDB('xrust_beze')
xb_db.createCollection("chats")
