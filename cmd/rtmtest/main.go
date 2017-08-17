package main

import (
	"errors"
	"flag"
	"os"

	"github.com/deliveroo/routemaster-client-go/integrationtest"
	"github.com/deliveroo/routemaster-client-go/integrationtest/suite"
)

var config struct {
	url       string
	rootToken string
}

func validateConfig() error {
	if config.url == "" {
		return errors.New("url is required")
	}
	if config.rootToken == "" {
		return errors.New("root-token is required")
	}
	return nil
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "help":
			flag.Usage()
			os.Exit(1)
		}
	}
	flag.StringVar(&config.url, "url", "", "url of routemaster bus")
	flag.StringVar(&config.rootToken, "root-token", "", "root token")

	flag.Parse()

	if err := validateConfig(); err != nil {
		flag.Usage()
		os.Exit(1)
	}

	runner := &integrationtest.Runner{
		URL:       config.url,
		RootToken: config.rootToken,
	}
	suite.ConfigureRunner(runner)
	runner.RunTests()
}
