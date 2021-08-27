package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/akinbezatoglu/survey-web-app/pkg/server"

	"github.com/sirupsen/logrus"
)

var (
	configFileFlag = flag.String("config.file", "config.yml", "Path to the configuration file.")
)

func init() {
	// Parse command-line flags
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	// Create server instance
	instance := server.NewInstance()

	// Interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		logrus.Infof("Received %s signal", <-c)
		instance.Shutdown()
	}()

	// Start server
	instance.Start(*configFileFlag)
}
