package runtime

import (
	"log/slog"
	"os"
	"strings"

	"xn--gckvb8fzb.com/switchyard/helpers/log"
	"xn--gckvb8fzb.com/switchyard/models/config"
)

var CONFIG_ENV_VAR = "SWITCHYARD_CONFIG"

var (
	Version string
	Commit  string
	Date    string
)

type Build struct {
	Version string
	Commit  string
	Date    string
}

type Runtime struct {
	Build  Build
	Logger *log.Logger
	Config *config.Config
}

func New(cfgstr string) (*Runtime, error) {
	var err error

	rt := new(Runtime)

	rt.Build.Version = Version
	rt.Build.Commit = Commit
	rt.Build.Date = Date

	if cfgstr == "" {
		if env, found := os.LookupEnv(CONFIG_ENV_VAR); found {
			cfgstr = env
		}
	}

	rt.Config, err = config.New(cfgstr)
	if err != nil {
		return nil, err
	}

	lvl := slog.LevelInfo
	if logcfg, err := rt.Config.Log(); err == nil {
		lvl = parseLevel(logcfg.Level)
	}
	rt.Logger = log.New(lvl)

	rt.Debug("Loaded configuration", "config", cfgstr)

	return rt, nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (rt *Runtime) Info(msg string, args ...interface{})    { rt.Logger.Info(msg, args...) }
func (rt *Runtime) Debug(msg string, args ...interface{})   { rt.Logger.Debug(msg, args...) }
func (rt *Runtime) Warning(msg string, args ...interface{}) { rt.Logger.Warning(msg, args...) }
func (rt *Runtime) Error(msg string, args ...interface{})   { rt.Logger.Error(msg, args...) }

func (rt *Runtime) NilOrDie(err error) {
	if err != nil {
		rt.Logger.Die(err.Error())
	}
}

func (rt *Runtime) Exit(code int) {
	rt.Debug("Ending runtime ...")
	os.Exit(code)
}
