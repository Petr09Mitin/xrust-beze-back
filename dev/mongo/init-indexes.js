db.auth("root", "root")
xb_db = db.getSiblingDB('xrust_beze')
xb_db.users.createIndex({"username": "text"})
xb_db.users.createIndex({"username": 1})

xb_db.reviews.createIndex({"user_id_to": 1, "created": 1})
