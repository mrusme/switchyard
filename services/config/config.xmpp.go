package config

import (
	glidesconfig "xn--gckvb8fzb.com/glides/services/config"
)

type XMPP struct {
	Server             string `koanf:"Server"`
	Username           string `koanf:"Username"`
	Password           string `koanf:"Password"`
	DirectTLS          bool   `koanf:"DirectTLS"`
	InsecureSkipVerify bool   `koanf:"InsecureSkipVerify"`
	Status             string `koanf:"Status"`
}

func GetXMPP(cfg *glidesconfig.Config) (x XMPP, err error) {
	_, err = cfg.Unmarshal("XMPP", &x)
	return x, err
}
