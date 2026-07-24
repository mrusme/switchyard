package config

import (
	"net/url"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"xn--gckvb8fzb.com/switchyard/errs"
)

type Log struct {
	Level string `koanf:"level"`
}

type Redis struct {
	Addrs      []string `koanf:"addrs"`
	Database   int      `koanf:"database"`
	Username   string   `koanf:"username"`
	Password   string   `koanf:"password"`
	Poolsize   int      `koanf:"poolsize"`
	MasterName string   `koanf:"master_name"`
}

type TLS struct {
	Enable bool   `koanf:"enable"`
	Cert   string `koanf:"cert"`
	Key    string `koanf:"key"`
}

type SMTP struct {
	Domain            string   `koanf:"domain"`
	AllowedIPs        []string `koanf:"allowed_ips"`
	Username          string   `koanf:"username"`
	Password          string   `koanf:"password"`
	ReadTimeout       string   `koanf:"read_timeout"`
	WriteTimeout      string   `koanf:"write_timeout"`
	MaxMessageBytes   int64    `koanf:"max_message_bytes"`
	MaxRecipients     int      `koanf:"max_recipients"`
	SMTPSAddress      string   `koanf:"smtps_address"`
	SubmissionAddress string   `koanf:"submission_address"`
	TLS               TLS      `koanf:"tls"`
}

type XMPP struct {
	Server             string `koanf:"server"`
	Username           string `koanf:"username"`
	Password           string `koanf:"password"`
	DirectTLS          bool   `koanf:"direct_tls"`
	InsecureSkipVerify bool   `koanf:"insecure_skip_verify"`
	Status             string `koanf:"status"`
}

type Config struct {
	cfgstr   string
	k        *koanf.Koanf
	provider koanf.Provider
}

func New(cfgstr string) (cfg *Config, err error) {
	cfg = new(Config)
	cfg.cfgstr = cfgstr

	var path string
	if path, err = cfg.parsePath(); err != nil {
		return nil, err
	}

	cfg.k = koanf.New(".")
	cfg.provider = file.Provider(path)
	if err = cfg.k.Load(cfg.provider, toml.Parser()); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) parsePath() (string, error) {
	u, err := url.Parse(cfg.cfgstr)
	if err != nil {
		return "", err
	}

	switch u.Scheme {
	case "":
		return cfg.cfgstr, nil
	case "file":
		if u.Host != "" {
			return u.Host + u.Path, nil
		}
		return u.Path, nil
	default:
		return "", errs.ErrConfigTypeUnsupported
	}
}

func (cfg *Config) Log() (l Log, err error) {
	err = cfg.k.Unmarshal("log", &l)
	return l, err
}

func (cfg *Config) Redis() (r Redis, err error) {
	err = cfg.k.Unmarshal("redis", &r)
	return r, err
}

func (cfg *Config) SMTP() (s SMTP, err error) {
	err = cfg.k.Unmarshal("smtp", &s)
	return s, err
}

func (cfg *Config) XMPP() (x XMPP, err error) {
	err = cfg.k.Unmarshal("xmpp", &x)
	return x, err
}
