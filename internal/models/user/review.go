package user

type Review struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Text     string `json:"text" bson:"text"`
	Rating   int    `json:"rating" bson:"rating" validate:"required,min=1,max=5"`
	UserIDBy string `json:"user_id_by" bson:"user_id_by" validate:"required"`
	UserIDTo string `json:"user_id_to" bson:"user_id_to" validate:"required"`
	Created  int64  `json:"created" bson:"created"`
	Updated  int64  `json:"updated" bson:"updated"`
}
