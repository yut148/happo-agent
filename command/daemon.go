package command

import (
	"crypto/tls"
	"fmt"
	"log"
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

	_ "net/http/pprof"

	"github.com/client9/reopen"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/secure"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/collect"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/model"
	"github.com/heartbeatsjp/happo-agent/util"
	"github.com/heartbeatsjp/happo-lib"
	"github.com/martini-contrib/binding"
)

// --- Struct
type daemonListener struct {
	Timeout        int
	MaxConnections int
	Port           string
	Handler        http.Handler
	PublicKey      string
	PrivateKey     string
}

// --- functions
func CmdDaemonWrapper(c *cli.Context) {
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
		envHappoUserId := os.Getenv("HAPPO_USER_ID")
		if envHappoUserId != "" {
			uid, err := strconv.Atoi(envHappoUserId)
			if err != nil {
				log.Print("HAPPO_USER_ID ", envHappoUserId)
				log.Fatal(err)
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
	m.Use(util.Logger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("public"))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}

// Daemon mode (agent mode)
func CmdDaemon(c *cli.Context) {

	fp, err := reopen.NewFileWriter(c.String("logfile"))
	if err != nil {
		fmt.Println(err)
	}
	log.SetOutput(fp)
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

	m.Get("/", func() string {
		return "OK"
	})

	util.CommandTimeout = time.Duration(c.Int("command-timeout"))
	model.MetricConfigFile = c.String("metric-config")

	m.Post("/proxy", binding.Json(happo_agent.ProxyRequest{}), model.Proxy)
	m.Post("/inventory", binding.Json(happo_agent.InventoryRequest{}), model.Inventory)
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), model.Monitor)
	m.Post("/metric", binding.Json(happo_agent.MetricRequest{}), model.Metric)
	m.Post("/metric/append", binding.Json(happo_agent.MetricAppendRequest{}), model.MetricAppend)
	m.Post("/metric/config/update", binding.Json(happo_agent.MetricConfigUpdateRequest{}), model.MetricConfigUpdate)
	m.Get("/metric/status", model.MetricDataBufferStatus)
	m.Get("/machine-state/", model.ListMachieState)
	m.Get("/machine-state/:key", model.GetMachineState)

	// Listener
	var lis daemonListener
	lis.Port = fmt.Sprintf(":%d", c.Int("port"))
	lis.Handler = m
	lis.Timeout = happo_agent.HTTP_TIMEOUT
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
	time_metrics := time.NewTicker(time.Minute).C
	for {
		select {
		case <-time_metrics:
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

	tls_config := &tls.Config{
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
	limit_listener := netutil.LimitListener(listener, l.MaxConnections)
	tls_listener := tls.NewListener(limit_listener, tls_config)

	http_config := &http.Server{
		TLSConfig:    tls_config,
		Addr:         l.Port,
		Handler:      l.Handler,
		ReadTimeout:  happo_agent.HTTP_TIMEOUT * time.Second,
		WriteTimeout: happo_agent.HTTP_TIMEOUT * time.Second,
	}

	return http_config.Serve(tls_listener)
}
