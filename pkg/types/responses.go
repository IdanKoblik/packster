package types

type HealthResponse struct {
	MySQLStatus string `json:"mysql"`
	RedisStatus string `json:"redis"`
}
