package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	"democart/database"
	"democart/server"
)

var (
	configFileFlag = flag.String("config", "", "config file path")
	dbURLFlag      = flag.String("db_url", "", "database url") // file override
	logLevelFlag   = flag.String("loglevel", "", "log level")  // file override
	version        = "<unknown>"
)

type HCLConfig struct {
	DBURL                   string   `hcl:"db_url"`
	ServerURL               string   `hcl:"server_url"`
	GracefulShutdownTimeout int      `hcl:"graceful_shutdown_timeout_sec"`
	WriteTimeout            int      `hcl:"write_timeout_sec"`
	ReadTimeout             int      `hcl:"read_timeout_sec"`
	IdleTimeout             int      `hcl:"idle_timeout_sec"`
	IDPPasswordSalt         string   `hcl:"idp_password_salt"`
	IDPClientID             string   `hcl:"idp_client_id"`
	IDPClientSecret         string   `hcl:"idp_client_secret"`
	LogLevel                string   `hcl:"loglevel"`
	DeveloperMode           bool     `hcl:"developer_mode"`
	ClientHosts             []string `hcl:"client_hosts"`
}

func main() {
	flag.Parse()
	config, err := parseAndValidateConfigFile(*configFileFlag, *dbURLFlag,
		*logLevelFlag)
	if err != nil {
		logrus.Fatalf("configuration error: %+v", err)
	}

	if err := run(context.Background(), config); err != nil {
		logrus.Fatalf("%+v", err)
	}
}

func run(ctx context.Context, c *Config) error {
	db, err := database.Connect(c.DBURL, nil)
	if err != nil {
		return err
	}

	logrus.Infof("starting democart version %q", version)
	s := server.New(db, &server.Config{
		IDPPasswordSalt: c.IDPPasswordSalt,
		IDPClientID:     c.IDPClientID,
		IDPClientSecret: c.IDPClientSecret,
		ServerURL:       c.ServerURL,
		DeveloperMode:   c.DeveloperMode,
		ClientHosts:     c.ClientHosts,
	})

	server := &http.Server{
		Addr:         c.ServerAddr,
		WriteTimeout: c.WriteTimeout,
		ReadTimeout:  c.ReadTimeout,
		IdleTimeout:  c.IdleTimeout,
		Handler:      s,
	}

	go func() {
		logrus.Infof("waiting for connections on %s", c.ServerURL.String())
		_ = server.ListenAndServe()
		// ignoring the possible error here
	}()

	interruptWaiter := make(chan os.Signal, 1)
	signal.Notify(interruptWaiter, os.Interrupt)
	<-interruptWaiter // block until interrupt signal received

	// set timeout incase something takes forever after interrupt
	ctx, cancel := context.WithTimeout(ctx, c.GracefulShutdownTimeout)
	defer cancel()

	go func() {
		logrus.Infof("shutting down...")
		_ = server.Shutdown(ctx)
		// ignoring the possible error here
	}()

	// wait for gracefull shutdown or canceled context
	<-ctx.Done()

	logrus.Infof("shut down")
	return nil
}

type Config struct {
	DBURL                   *url.URL
	ServerURL               *url.URL
	ServerAddr              string
	GracefulShutdownTimeout time.Duration
	WriteTimeout            time.Duration
	ReadTimeout             time.Duration
	IdleTimeout             time.Duration
	IDPPasswordSalt         string
	IDPClientID             string
	IDPClientSecret         string
	LogLevel                logrus.Level
	DeveloperMode           bool
	ClientHosts             []*url.URL
}

func parseAndValidateConfigFile(configFile, dbURL, loglevel string) (*Config,
	error) {
	hclBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	hclconfig := &HCLConfig{}
	if err := hcl.Unmarshal(hclBytes, hclconfig); err != nil {
		return nil, errs.Wrap(err)
	}

	config, err := hclconfig.validate(dbURL, loglevel)
	if err != nil {
		return nil, err
	}

	logrus.SetLevel(config.LogLevel)
	return config, nil
}

func (hclconfig *HCLConfig) validate(dbURLOverride, loglevelOverride string) (
	*Config, error) {
	if hclconfig.DBURL == "" && dbURLOverride == "" {
		return nil, errs.New("db_url unconfigured")
	}
	if hclconfig.ServerURL == "" {
		return nil, errs.New("server_url unconfigured")
	}
	if hclconfig.GracefulShutdownTimeout == 0 {
		return nil, errs.New("graceful_shutdown_timeout_sec unconfigured")
	}
	if hclconfig.WriteTimeout == 0 {
		return nil, errs.New("write_sec unconfigured")
	}
	if hclconfig.ReadTimeout == 0 {
		return nil, errs.New("read_sec unconfigured")
	}
	if hclconfig.IdleTimeout == 0 {
		return nil, errs.New("idle_sec unconfigured")
	}
	if hclconfig.IDPPasswordSalt == "" {
		return nil, errs.New("idp_password_salt unconfigured")
	}
	if hclconfig.IDPClientID == "" {
		return nil, errs.New("idp_client_id unconfigured")
	}
	if hclconfig.IDPClientSecret == "" {
		return nil, errs.New("idp_client_secret unconfigured")
	}
	if hclconfig.LogLevel == "" && loglevelOverride == "" {
		return nil, errs.New("loglevel unconfigured")
	}
	if len(hclconfig.ClientHosts) == 0 {
		return nil, errs.New("client_hosts unconfigured")
	}

	if dbURLOverride != "" {
		hclconfig.DBURL = dbURLOverride
	}
	dbURL, err := url.Parse(hclconfig.DBURL)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(hclconfig.ServerURL)
	if err != nil {
		return nil, err
	}
	serverAddr := ":" + serverURL.Port()
	if serverAddr == ":" {
		return nil, errs.New("server_url misconfigured. a port is necessary")
	}

	grace := time.Second * time.Duration(hclconfig.GracefulShutdownTimeout)
	write := time.Second * time.Duration(hclconfig.WriteTimeout)
	read := time.Second * time.Duration(hclconfig.ReadTimeout)
	idle := time.Second * time.Duration(hclconfig.IdleTimeout)

	if loglevelOverride != "" {
		hclconfig.LogLevel = loglevelOverride
	}
	loglevel, err := logrus.ParseLevel(hclconfig.LogLevel)
	if err != nil {
		return nil, err
	}

	clientHosts := make([]*url.URL, 0, len(hclconfig.ClientHosts))
	for _, h := range hclconfig.ClientHosts {
		host, err := url.Parse(h)
		if err != nil {
			return nil, err
		}
		clientHosts = append(clientHosts, host)
	}

	c := &Config{
		DBURL:                   dbURL,
		ServerURL:               serverURL,
		ServerAddr:              serverAddr,
		GracefulShutdownTimeout: grace,
		WriteTimeout:            write,
		ReadTimeout:             read,
		IdleTimeout:             idle,
		IDPPasswordSalt:         hclconfig.IDPPasswordSalt,
		IDPClientID:             hclconfig.IDPClientID,
		IDPClientSecret:         hclconfig.IDPClientSecret,
		LogLevel:                loglevel,
		DeveloperMode:           hclconfig.DeveloperMode,
		ClientHosts:             clientHosts,
	}
	return c, nil
}
