package config

type Config struct {
	Secret string `yaml:"secret"`
	FileUploadLimit int `yaml:"file_upload_limit,omitempty"`

	Sql     PgsqlConfig    `yaml:"sql"`
	Gitlab  *GitlabConfig  `yaml:"gitlab"`
	Storage StorageConfig  `yaml:"storage"`
}

type StorageConfig struct {
	Path string `yaml:"path"`
}

type PgsqlConfig struct {
	Host 	 string  `yaml:"host"`
	Port 	 uint16  `yaml:"port"`
	DB 		 string  `yaml:"db"`
	User 	 string  `yaml:"user"`
	Password string	 `yaml:"password"`
	SSL  	 bool    `yaml:"ssl"`
}

type GitlabConfig struct {
	Hosts map[string]GitlabHost `yaml:"hosts"`
}

type GitlabHost struct {
	ApplicationId string `yaml:"id"`
	Secret		  string `yaml:"secret"`
}
