package types

type HealthResponse struct {
	MongoStatus string `json:"mongo"`
	RedisStatus string `json:"redis"`
}
