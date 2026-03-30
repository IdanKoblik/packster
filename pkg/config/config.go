package config

type Config struct {
	FileUploadLimit int `yaml:"file_upload_limit,omitempty"`

	MySQL   MySQLConfig   `yaml:"mysql"`
	Redis   RedisConfig   `yaml:"redis"`
	Metrics MetricsConfig `yaml:"metrics,omitempty"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db,omitempty"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type MetricsConfig struct {
	Addr string `yaml:"addr,omitempty"`
}
