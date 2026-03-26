package config

type Config struct {
	FileUploadLimit int `yaml:"file_upload_limit,omitempty"`

	Mongo   MongoConfig   `yaml:"mongo"`
	Redis   RedisConfig   `yaml:"redis"`
	Metrics MetricsConfig `yaml:"metrics,omitempty"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db,omitempty"`
}

type MongoConfig struct {
	ConnectionString  string `yaml:"connection_string"`
	Database          string `yaml:"database"`
	TokenCollection   string `yaml:"token_collection"`
	ProductCollection string `yaml:"product_collection"`
}

type MetricsConfig struct {
	Addr string `yaml:"addr,omitempty"`
}
