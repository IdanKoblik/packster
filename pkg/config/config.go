package config

type Config struct {
	FileUploadLimit int `yaml:"file_upload_limit,omitempty"`

	Sql PgsqlConfig `yaml:"sql"`
	Gitlab *GitlabConfig `yaml:"gitlab,omitempty"`
}

type PgsqlConfig struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
	DB string `yaml:"db"`
	User string `yaml:"user"`
	Password string `yaml:"password"`
	SSL  bool   `yaml:"ssl"`
}

type GitlabConfig struct {
	Host          string `yaml:"host"`
	ApplicationId string `yaml:"application_id"`
	Secret        string `yaml:"secret"`
}
