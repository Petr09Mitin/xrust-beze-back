package chat_models

import "go.mongodb.org/mongo-driver/v2/bson"

type BSONChannel struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	UserIDs []string      `bson:"user_ids"`
	Created int64         `bson:"created"`
	Updated int64         `bson:"updated"`
}

func (c *BSONChannel) ToChannel() Channel {
	return Channel{
		ID:      c.ID.Hex(),
		UserIDs: c.UserIDs,
		Created: c.Created,
		Updated: c.Updated,
	}
}
