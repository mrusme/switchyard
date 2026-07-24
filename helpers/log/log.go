package log

import (
	"fmt"
	"log/slog"
	"os"
)

type Logger struct {
	engine *slog.Logger
}

func New(lvl slog.Level) *Logger {
	log := new(Logger)
	log.engine = slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: lvl,
		}),
	)

	return log
}

func (logger *Logger) Error(msg string, args ...interface{}) {
	logger.engine.Error(msg, args...)
}

func (logger *Logger) Warning(msg string, args ...interface{}) {
	logger.engine.Warn(msg, args...)
}

func (logger *Logger) Info(msg string, args ...interface{}) {
	logger.engine.Info(msg, args...)
}

func (logger *Logger) Debug(msg string, args ...any) {
	logger.engine.Debug(msg, args...)
}

func (logger *Logger) Errorf(msg string, args ...interface{}) {
	logger.engine.Error(fmt.Sprintf(msg, args...))
}

func (logger *Logger) Warningf(msg string, args ...interface{}) {
	logger.engine.Warn(fmt.Sprintf(msg, args...))
}

func (logger *Logger) Infof(msg string, args ...interface{}) {
	logger.engine.Info(fmt.Sprintf(msg, args...))
}

func (logger *Logger) Debugf(msg string, args ...interface{}) {
	logger.engine.Debug(fmt.Sprintf(msg, args...))
}

func (logger *Logger) NilOrDie(err error, msg string) {
	if err != nil {
		logger.engine.Error(msg+"\n", "error", err)
		os.Exit(1)
	}
}

func (logger *Logger) Die(msg string, args ...interface{}) {
	logger.engine.Error(msg, args...)
	os.Exit(1)
}

type Asynq struct {
	engine *slog.Logger
}

func (logger *Logger) Asynq() *Asynq {
	return &Asynq{engine: logger.engine}
}

func (a *Asynq) Debug(args ...interface{}) { a.engine.Debug(fmt.Sprint(args...)) }
func (a *Asynq) Info(args ...interface{})  { a.engine.Info(fmt.Sprint(args...)) }
func (a *Asynq) Warn(args ...interface{})  { a.engine.Warn(fmt.Sprint(args...)) }
func (a *Asynq) Error(args ...interface{}) { a.engine.Error(fmt.Sprint(args...)) }
func (a *Asynq) Fatal(args ...interface{}) {
	a.engine.Error(fmt.Sprint(args...))
	os.Exit(1)
}
