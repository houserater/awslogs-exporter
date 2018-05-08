package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/houserater/awslogs-exporter/log"
)

const (
	defaultListenAddress    = ":9223"
	defaultAwsRegion        = ""
	defaultMetricsPath      = "/metrics"
	defaultGroupFilter      = ""
	defaultLogHistory       = 3600
	defaultLogJSONFormat    = ""
	defaultDebug            = false
)

// Cfg is the global configuration
var cfg *config

// Parse global config
func parse(args []string) error {
	return cfg.parse(args)
}

// Config represents an app configuration
type config struct {
	fs *flag.FlagSet

	listenAddress    string
	awsRegion        string
	metricsPath      string
	groupFilter      string
	logHistory       int64
	logJSONFormat    string
	debug            bool
}

// init will load all the flags
func init() {
	cfg = new()
}

// New returns an initialized config
func new() *config {
	c := &config{
		fs: flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
	}

	c.fs.StringVar(
		&c.listenAddress, "web.listen-address", defaultListenAddress, "Address to listen on")

	c.fs.StringVar(
		&c.awsRegion, "aws.region", defaultAwsRegion, "The AWS region to get metrics from")

	c.fs.StringVar(
		&c.groupFilter, "aws.log-prefix", defaultGroupFilter, "Name prefix used to filter the log group names")

	c.fs.Int64Var(
		&c.logHistory, "aws.log-history", defaultLogHistory, "Number of seconds of previous log events to search")

	c.fs.StringVar(
		&c.logJSONFormat, "aws.log-json-format", defaultLogJSONFormat, "Parse log lines as JSON and output in format (i.e. {name}: {message})")

	c.fs.StringVar(
		&c.metricsPath, "web.telemetry-path", defaultMetricsPath, "The path where metrics will be exposed")

	c.fs.BoolVar(
		&c.debug, "debug", defaultDebug, "Run exporter in debug mode")

	return c
}

// parse parses the flags for configuration
func (c *config) parse(args []string) error {
	log.Debugf("Parsing flags...")

	if err := c.fs.Parse(args); err != nil {
		return err
	}

	if len(c.fs.Args()) != 0 {
		return fmt.Errorf("Invalid command line arguments. Help: %s -h", os.Args[0])
	}

	if c.awsRegion == "" {
		return fmt.Errorf("An aws region is required")
	}

	if c.groupFilter != defaultGroupFilter {
		log.Warnf("Filtering cluster metrics by: %s", c.groupFilter)
	}

	return nil
}
