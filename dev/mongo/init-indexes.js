db.auth("root", "root")
xb_db = db.getSiblingDB('xrust_beze')
xb_db.users.createIndex({"username": "text"})
xb_db.users.createIndex({"username": 1})
