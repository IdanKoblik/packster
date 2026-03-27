package types

type ApiToken struct {
	Token string `json:"token" bson:"_id"`
	Admin bool   `json:"admin" bson:"-"`
}
