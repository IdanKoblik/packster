package config

import (
	"fmt"
	"net/url"
)
type Config struct {
	FileUploadLimit int `yaml:"file_upload_limit,omitempty"`

	Sql MysqlConfig `yaml:"sql"`
}

type MysqlConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

func (c MysqlConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		c.Username,
		url.QueryEscape(c.Password),
		c.Host,
		c.DB,
	)
}
