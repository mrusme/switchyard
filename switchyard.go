package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"xn--gckvb8fzb.com/glides/runtime"
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
		Services: runtime.Services{
			Dispatch: true,
		},
	})
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	err = rt.Startup()
	rt.NilOrDie(err)

	// ---[ XMPP ]------------------------------------------------------------- //
	xmp, err := xmpp.New(rt)
	rt.NilOrDie(err)
	rt.NilOrDie(xmp.Run())

	// ---[ DISPATCH ]--------------------------------------------------------- //
	disp, err := dispatch.New(rt)
	rt.NilOrDie(err)

	// ---[ WORKER ]----------------------------------------------------------- //
	wrk, err := worker.New(rt, xmp)
	rt.NilOrDie(err)
	go wrk.Run()

	// ---[ SMTP ]------------------------------------------------------------- //
	srv, err := smtp.New(rt, disp)
	rt.NilOrDie(err)
	go srv.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	srv.Shutdown()
	wrk.Shutdown()
	xmp.Shutdown()

	rt.Exit(0)
}
