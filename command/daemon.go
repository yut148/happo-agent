package command

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime/pprof"
	"time"

	"golang.org/x/net/netutil"

	_ "net/http/pprof"

	"github.com/codegangsta/cli"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/secure"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/collect"
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
	fmt.Println(args[0])
	for {
		cmd := exec.Command(args[0], args[1:]...)
		stdout, _ := cmd.StdoutPipe()
		cmd.Start()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		cmd.Wait()
	}
}

// Daemon mode (agent mode)
func CmdDaemon(c *cli.Context) {
	m := martini.Classic()
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

	m.Get("/", func() string {
		return "OK"
	})

	util.CommandTimeout = time.Duration(c.Int("command-timeout"))
	model.MetricConfigFile = c.String("metric-config")

	m.Post("/proxy", binding.Json(happo_agent.ProxyRequest{}), model.Proxy)
	m.Post("/inventory", binding.Json(happo_agent.InventoryRequest{}), model.Inventory)
	m.Post("/monitor", binding.Json(happo_agent.MonitorRequest{}), model.Monitor)
	m.Post("/metric", binding.Json(happo_agent.MetricRequest{}), model.Metric)
	m.Post("/metric/config/update", binding.Json(happo_agent.MetricConfigUpdateRequest{}), model.MetricConfigUpdate)
	m.Get("/metric/status", model.MetricDataBufferStatus)

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
