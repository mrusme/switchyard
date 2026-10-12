package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"xn--gckvb8fzb.com/glides/runtime"
	glidesdispatch "xn--gckvb8fzb.com/glides/services/dispatch"
	"xn--gckvb8fzb.com/switchyard/services/dispatch"
	"xn--gckvb8fzb.com/switchyard/smtp"
	"xn--gckvb8fzb.com/switchyard/worker"
	"xn--gckvb8fzb.com/switchyard/xmpp"
)

var (
	Version string
	Commit  string
	Date    string
)

const CONFIG_ENV_VAR string = "SWITCHYARD_CONFIG"

var (
	flagCfgstr  string
	flagVersion bool
)

func init() {
	flag.StringVar(&flagCfgstr, "c", "file:///etc/switchyard.toml", "configuration string")
	flag.BoolVar(&flagVersion, "v", false, "Print version information and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use: %s [-opts]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if flagVersion {
		fmt.Printf("switchyard %s\nCommit: %s\nBuild date: %s\n",
			Version, Commit, Date)
		os.Exit(0)
	}

	cfgstr := flagCfgstr
	if cfgstr == "" {
		if env, found := os.LookupEnv(CONFIG_ENV_VAR); found {
			cfgstr = env
		}
	}

	rt, err := runtime.New(runtime.Opts{
		Cfgstr:  cfgstr,
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	})
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	if _, err = glidesdispatch.Register(rt); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	err = rt.Startup()
	rt.NilOrDie(err)

	xmp, err := xmpp.New(rt)
	rt.NilOrDie(err)

	disp, err := dispatch.New(rt)
	rt.NilOrDie(err)

	wrk, err := worker.New(rt, xmp)
	rt.NilOrDie(err)

	srv, err := smtp.New(rt, disp)
	rt.NilOrDie(err)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err = rt.Serve(ctx, xmp, wrk, srv); err != nil {
		rt.Exit(1)
	}

	rt.Exit(0)
}
