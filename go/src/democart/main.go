package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	"democart/config"
	"democart/idp"
	api "democart/server"
)

//
// look in config/validate.go for configuration flag/env values
//

// version is set at build time by command line arg and included with logs
var version = "<unknown>"

func main() {
	err := run()
	if err != nil {
		logrus.Fatalf("democart: %+v", err)
	}
}

func run() error {
	// pulls in flag files, flag values, and environment variables
	conf, err := config.Parse()
	if err != nil {
		return errs.New("configuration error: %+v", err)
	}

	// set loglevel defined in config
	logrus.SetLevel(conf.LogLevel)
	conf.SetVersion(version)

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// stay alive for all goroutines to finish
	wg := sync.WaitGroup{}

	//// initialize the metric server
	//metricClient, metricServer, metricMiddleware, err := prometheus.NewHTTPServer(conf)
	//if err != nil {
	//	return err
	//}
	//defer metricClient.Close()

	// initialize the idp server
	idpClient, idpServer, err := idp.NewHTTPServer(conf)
	if err != nil {
		return err
	}
	defer idpClient.Close()

	// initialize the api server
	//apiClient, apiServer, err := api.NewHTTPServer(conf, metricMiddleware)
	apiClient, apiServer, err := api.NewHTTPServer(conf)
	if err != nil {
		return err
	}
	defer apiClient.Close()

	//// service 1 - start the metric server
	//wg.Add(1)
	//go gracefullyServe(ctx, &wg, metricServer, conf.GracefulShutdownTimeout)

	// service 2 - start the idp server
	wg.Add(1)
	go gracefullyServe(ctx, &wg, idpServer, conf.GracefulShutdownTimeout)

	// service 3 - start the api server
	logrus.Infof("starting democart version %q", version)
	wg.Add(1)
	go gracefullyServe(ctx, &wg, apiServer, conf.GracefulShutdownTimeout)

	// listen for C-c interrupt
	interruptWaiter := make(chan os.Signal, 1)
	signal.Notify(interruptWaiter, os.Interrupt)
	<-interruptWaiter // block until interrupt signal received

	cancel() // cancel context and let gracefullyServes spin down
	wg.Wait()

	logrus.Info("shut down")
	return nil
}

func gracefullyServe(ctx context.Context, wg *sync.WaitGroup, s *http.Server,
	shutdownTimeout time.Duration) {

	defer wg.Done()
	go func() {
		// start the server
		if err := s.ListenAndServe(); err != nil {
			logrus.Errorf("%s", err)
		}
	}()

	<-ctx.Done() // blocks until the context is cancelled
	logrus.Info("shutting down...")

	// set timeout with a new context (the last one has been canceled) incase
	// something takes forever after interrupt
	shutdownCtx, cancel := context.WithTimeout(context.Background(),
		shutdownTimeout)
	defer cancel()

	go func() {
		_ = s.Shutdown(shutdownCtx)
		// ignore "Error shutting down server: context canceled"
	}()

	logrus.Debug("server gracefully stopped")
}
