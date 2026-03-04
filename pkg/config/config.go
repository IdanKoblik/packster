package config

type Config struct {
	FileUploadLimit int `yaml:"file_upload_limit"`
	Sql PgsqlConfig `yaml:"sql"`
}

type PgsqlConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Addr     string `yaml:"addr"`
	Database string `yaml:"database"`
}
