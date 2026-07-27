package config

import (
	glidesconfig "xn--gckvb8fzb.com/glides/services/config"
)

type TLS struct {
	Enable bool   `koanf:"Enable"`
	Cert   string `koanf:"Cert"`
	Key    string `koanf:"Key"`
}

type SMTP struct {
	Domain            string   `koanf:"Domain"`
	AllowedIPs        []string `koanf:"AllowedIPs"`
	Username          string   `koanf:"Username"`
	Password          string   `koanf:"Password"`
	ReadTimeout       string   `koanf:"ReadTimeout"`
	WriteTimeout      string   `koanf:"WriteTimeout"`
	MaxMessageBytes   int64    `koanf:"MaxMessageBytes"`
	MaxRecipients     int      `koanf:"MaxRecipients"`
	SMTPSAddress      string   `koanf:"SMTPSAddress"`
	SubmissionAddress string   `koanf:"SubmissionAddress"`
	TLS               TLS      `koanf:"TLS"`
}

func GetSMTP(cfg *glidesconfig.Config) (s SMTP, err error) {
	_, err = cfg.Unmarshal("SMTP", &s)
	return s, err
}
