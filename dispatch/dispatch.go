package dispatch

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"xn--gckvb8fzb.com/switchyard/models/config"
	"xn--gckvb8fzb.com/switchyard/models/message"
	"xn--gckvb8fzb.com/switchyard/runtime"
)

const TaskDeliver = "deliver"

type Dispatch struct {
	rt  *runtime.Runtime
	cfg config.Redis
	ac  *asynq.Client
}

func New(rt *runtime.Runtime) (disp *Dispatch, err error) {
	disp = new(Dispatch)
	disp.rt = rt

	if disp.cfg, err = rt.Config.Redis(); err != nil {
		return nil, err
	}

	return disp, nil
}

func (disp *Dispatch) Startup() error {
	disp.ac = asynq.NewClient(RedisConnOpt(disp.cfg))
	return nil
}

func (disp *Dispatch) Shutdown() error {
	if disp.ac != nil {
		return disp.ac.Close()
	}
	return nil
}

func (disp *Dispatch) Deliver(m *message.Message) (err error) {
	payload, err := json.Marshal(m)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TaskDeliver, payload,
		asynq.MaxRetry(10),
		asynq.Timeout(5*time.Minute),
	)

	info, err := disp.ac.Enqueue(task)
	if err != nil {
		return err
	}

	disp.rt.Debug("enqueued",
		"id", info.ID, "recipients", m.Recipients)

	return nil
}

func RedisConnOpt(cfg config.Redis) asynq.RedisConnOpt {
	if len(cfg.Addrs) > 1 {
		return asynq.RedisClusterClientOpt{
			Addrs:    cfg.Addrs,
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	if cfg.MasterName != "" {
		return asynq.RedisFailoverClientOpt{
			MasterName:    cfg.MasterName,
			SentinelAddrs: cfg.Addrs,
			DB:            cfg.Database,
			Username:      cfg.Username,
			Password:      cfg.Password,
			PoolSize:      cfg.Poolsize,
		}
	}

	var addr string
	if len(cfg.Addrs) == 1 {
		addr = cfg.Addrs[0]
	}

	return asynq.RedisClientOpt{
		Addr:     addr,
		DB:       cfg.Database,
		Username: cfg.Username,
		Password: cfg.Password,
		PoolSize: cfg.Poolsize,
	}
}
