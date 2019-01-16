package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"crypto/rand"

	"github.com/Sirupsen/logrus"
)

var logger *logrus.Logger

func run(args []string) int {
	bindAddress := flag.String("ip", "0.0.0.0", "IP address to bind")
	listenPort := flag.Int("port", 25478, "port number to listen on")
	logLevelFlag := flag.String("loglevel", "info", "logging level")
	flag.Parse()
	serverRoot := flag.Arg(0)
	if len(serverRoot) == 0 {
		flag.Usage()
		return 2
	}
	if logLevel, err := logrus.ParseLevel(*logLevelFlag); err != nil {
		logrus.WithError(err).Error("failed to parse logging level, so set to default")
	} else {
		logger.Level = logLevel
	}
	token := *tokenFlag
	if token == "" {
		count := 10
		b := make([]byte, count)
		if _, err := rand.Read(b); err != nil {
			logger.WithError(err).Fatal("could not generate token")
			return 1
		}
		token = fmt.Sprintf("%x", b)
		logger.WithField("token", token).Warn("token generated")
	}
	logger.WithFields(logrus.Fields{
		"ip":           *bindAddress,
		"port":         *listenPort,
		"root":         serverRoot,
	}).Info("start listening")
	server := NewServer(serverRoot)
	http.Handle("/snaps/", server)
	http.ListenAndServe(fmt.Sprintf("%s:%d", *bindAddress, *listenPort), nil)
	return 0
}

func main() {
	logger = logrus.New()
	logger.Info("starting up simple-upload-server")

	result := run(os.Args)
	os.Exit(result)
}
