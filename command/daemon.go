package command

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/net/netutil"

	"github.com/client9/reopen"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/secure"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/model"
	"github.com/heartbeatsjp/happo-agent/util"
	"github.com/martini-contrib/binding"
)

// --- Struct
type daemonListener struct {
	Timeout        int //second
	MaxConnections int
	Port           string
	Handler        http.Handler
	PublicKey      string
	PrivateKey     string
}

// --- functions

// CmdDaemonWrapper implements subcommand `daemon`
func CmdDaemonWrapper(c *cli.Context) {
	log := util.HappoAgentLogger()

	args := os.Args
	args[1] = "_daemon"
	started := []time.Time{}

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt)
	signal.Notify(sigTerm, syscall.SIGTERM)
	var cmd *exec.Cmd

	go func() {
		<-sigTerm
		cmd.Process.Kill()
		os.Exit(1)
	}()

	sigHup := make(chan os.Signal, 1)
	signal.Notify(sigHup, syscall.SIGHUP)
	go func() {
		for {
			<-sigHup
			cmd.Process.Signal(syscall.SIGHUP)
		}
	}()

	for {
		cmd = exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		envHappoUserID := os.Getenv("HAPPO_USER_ID")
		if envHappoUserID != "" {
			uid, err := strconv.Atoi(envHappoUserID)
			if err != nil {
				log.Fatal("HAPPO_USER_ID ", envHappoUserID, err)
			}
			cmd.SysProcAttr = &syscall.SysProcAttr{}
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid)}
		}
		started = append(started, time.Now())

		cmd.Start()
		cmd.Wait()

		if len(started) > 10 && time.Now().Add(-30*time.Second).Before(started[len(started)-5]) {
			log.Fatal("Restarted too fast. Abort!")
			os.Exit(1)
		}
	}
}

// custom martini.Classic() for change change martini.Logger() to util.Logger()
func customClassic() *martini.ClassicMartini {
	/*
		- remove martini.Logging()
		- add happo_agent.martini_util.Logging()
	*/
	r := martini.NewRouter()
	m := martini.New()
	m.Use(util.MartiniCustomLogger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("public"))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	classic := new(martini.ClassicMartini)
	classic.Martini = m
	classic.Router = r
	return classic
}

// CmdDaemon implements subcommand `_daemon`
func CmdDaemon(c *cli.Context) {
	log := util.HappoAgentLogger()

	fp, err := reopen.NewFileWriter(c.String("logfile"))
	if err != nil {
		fmt.Println(err)
	}
	log.Info(fmt.Sprintf("switch log.Out to %s", c.String("logfile")))
	if !util.Production {
		log.Warn("MARTINI_ENV is not production. LogLevel force to debug")
		util.SetLogLevel(util.HappoAgentLogLevelDebug)
	}

	log.Out = fp
	sigHup := make(chan os.Signal, 1)
	signal.Notify(sigHup, syscall.SIGHUP)
	go func() {
		for {
			<-sigHup
			fp.Reopen()
		}
	}()

	m := customClassic()
	m.Use(render.Renderer())
	m.Use(util.ACL(c.StringSlice("allowed-hosts")))
	m.Use(
		secure.Secure(secure.Options{
			SSLRedirect:      true,
			DisableProdCheck: true,
		}))
	m.Use(util.MartiniRequestStatus())

	// CPU Profiling
	if c.String("cpu-profile") != "" {
		cpuprofile := c.String("cpu-profile")
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		cpuprof := make(chan os.Signal, 1)
		signal.Notify(cpuprof, os.Interrupt)
		go func() {
			for sig := range cpuprof {
				log.Printf("captured %v, stopping profiler and exiting...", sig)
				pprof.StopCPUProfile()
				os.Exit(1)
			}
		}()
	}

	dbfile := c.String("dbfile")
	db.Open(dbfile)
	defer db.Close()
	db.MetricsMaxLifetimeSeconds = c.Int64("metrics-max-lifetime-seconds")
	db.MachineStateMaxLifetimeSeconds = c.Int64("machine-state-max-lifetime-seconds")

	model.SetProxyTimeout(c.Int64("proxy-timeout-seconds"))

	model.AppVersion = c.App.Version
	m.Get("/", func() string {
		return "OK"
	})

	util.CommandTimeout = time.Duration(c.Int("command-timeout"))
	model.MetricConfigFile = c.String("metric-config")

	model.ErrorLogIntervalSeconds = c.Int64("error-log-interval-seconds")
	model.NagiosPluginPaths = c.String("nagios-plugin-paths")
	collect.SensuPluginPaths = c.String("sensu-plugin-paths")

	m.Post("/proxy", binding.Json(halib.ProxyRequest{}), model.Proxy)
	m.Post("/inventory", binding.Json(halib.InventoryRequest{}), model.Inventory)
	m.Post("/monitor", binding.Json(halib.MonitorRequest{}), model.Monitor)
	m.Post("/metric", binding.Json(halib.MetricRequest{}), model.Metric)
	m.Post("/metric/append", binding.Json(halib.MetricAppendRequest{}), model.MetricAppend)
	m.Post("/metric/config/update", binding.Json(halib.MetricConfigUpdateRequest{}), model.MetricConfigUpdate)
	m.Get("/metric/status", model.MetricDataBufferStatus)
	m.Get("/status", model.Status)
	m.Get("/status/request", model.RequestStatus)
	m.Get("/machine-state/", model.ListMachieState)
	m.Get("/machine-state/:key", model.GetMachineState)

	// Listener
	var lis daemonListener
	lis.Port = fmt.Sprintf(":%d", c.Int("port"))
	lis.Handler = m
	lis.Timeout = halib.DefaultServerHTTPTimeout
	if lis.Timeout < int(c.Int64("proxy-timeout-seconds")) {
		lis.Timeout = int(c.Int64("proxy-timeout-seconds"))
	}
	if lis.Timeout < c.Int("command-timeout") {
		lis.Timeout = c.Int("command-timeout")
	}
	lis.MaxConnections = c.Int("max-connections")
	lis.PublicKey = c.String("public-key")
	lis.PrivateKey = c.String("private-key")
	go func() {
		err := lis.listenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Metric collect timer
	timeMetrics := time.NewTicker(time.Minute).C
	for {
		select {
		case <-timeMetrics:
			err := collect.Metrics(c.String("metric-config"))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// HTTPS Listener
func (l *daemonListener) listenAndServe() error {
	var err error

	cert := make([]tls.Certificate, 1)
	cert[0], err = tls.LoadX509KeyPair(l.PublicKey, l.PrivateKey)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			// tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		NextProtos:               []string{"http/1.1"},
		Certificates:             cert,
	}

	listener, err := net.Listen("tcp", l.Port)
	if err != nil {
		return err
	}
	limitListener := netutil.LimitListener(listener, l.MaxConnections)
	tlsListener := tls.NewListener(limitListener, tlsConfig)

	httpConfig := &http.Server{
		TLSConfig:    tlsConfig,
		Addr:         l.Port,
		Handler:      l.Handler,
		ReadTimeout:  time.Duration(l.Timeout) * time.Second,
		WriteTimeout: time.Duration(l.Timeout) * time.Second,
	}

	return httpConfig.Serve(tlsListener)
}
