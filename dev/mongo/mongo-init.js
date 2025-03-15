use admin
db.auth("root", "root")
db.createUser(
    {
        user: "admin",
        pwd: "admin",
        roles: [
            {
                role: "readWrite",
                db: "xrust_beze"
            }
        ]
    }
);
