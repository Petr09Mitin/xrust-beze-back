db.auth("admin", "admin")
xb_db = db.getSiblingDB('xrust_beze')
xb_db.messages.deleteMany({})
xb_db.channels.deleteMany({})
xb_db.users.deleteMany({})

