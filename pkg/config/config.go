package config

type Config struct {
	FileUploadLimit int `yaml:"file_upload_limit,omitempty"`
	Sql PgsqlConfig `yaml:"sql"`
	Redis RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
	Password string `yaml:"password,omitempty"`
}

type PgsqlConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Addr     string `yaml:"addr"`
	Database string `yaml:"database"`
}
