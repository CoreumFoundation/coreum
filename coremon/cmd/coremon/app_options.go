package main

import (
	cli "github.com/jawher/mow.cli"
)

// initGlobalOptions defines some global CLI options, that are useful for most parts of the app.
// Before adding option to there, consider moving it into the actual Cmd.
func initGlobalOptions(
	envName **string,
) {
	*envName = app.String(cli.StringOpt{
		Name:   "e env",
		Desc:   "The environment name this app runs in. Used for metrics and error reporting.",
		EnvVar: "COREMON_ENV",
		Value:  "local",
	})
}

// initLoggerOptions sets options for the global Zap logger.
func initLoggerOptions(
	logFormat **string,
	logVerbose **string,
) {
	*logFormat = app.String(cli.StringOpt{
		Name:   "log-format",
		Desc:   "Format of log output: console | json",
		EnvVar: "COREMON_LOG_FORMAT",
		Value:  "json",
	})

	*logVerbose = app.String(cli.StringOpt{
		Name:   "v verbose",
		Desc:   "Turns on verbose logging.",
		EnvVar: "COREMON_LOG_VERBOSE",
		Value:  "false",
	})
}

func initCosmosOptions(
	c *cli.Cmd,
	chainID **string,
	tendermintRPC **string,
	parallelBlockFetchJobs **int,
) {
	*chainID = c.String(cli.StringOpt{
		Name:   "chain-id",
		Desc:   "Specify Chain ID of the network.",
		EnvVar: "COREMON_CHAIN_ID",
		Value:  "coredev",
	})

	*tendermintRPC = c.String(cli.StringOpt{
		Name:   "tendermint-rpc",
		Desc:   "Tendermint RPC endpoint",
		EnvVar: "COREMON_TENDERMINT_RPC",
		Value:  "http://localhost:26657",
	})

	*parallelBlockFetchJobs = c.Int(cli.IntOpt{
		Name:   "parallel-fetch-jobs",
		Desc:   "Number of sumultaneous jobs fetching the blocks from the RPC. Change only if need to access historical.",
		EnvVar: "COREMON_BLOCK_FETCH_JOBS",
		Value:  1,
	})
}

// initStatsdOptions sets options for StatsD metrics.
func initStatsdOptions(
	c *cli.Cmd,
	statsdPrefix **string,
	statsdAddr **string,
	statsdEnabled **string,
) {
	*statsdPrefix = c.String(cli.StringOpt{
		Name:   "statsd-prefix",
		Desc:   "Specify StatsD compatible metrics prefix.",
		EnvVar: "COREMON_STATSD_PREFIX",
		Value:  "coremon",
	})

	*statsdAddr = c.String(cli.StringOpt{
		Name:   "statsd-addr",
		Desc:   "UDP address of a StatsD compatible metrics aggregator.",
		EnvVar: "COREMON_STATSD_ADDR",
		Value:  "localhost:8125",
	})

	*statsdEnabled = c.String(cli.StringOpt{
		Name:   "statsd-enabled",
		Desc:   "Allows to disable StatsD reporting component.",
		EnvVar: "COREMON_STATSD_ENABLED",
		Value:  "true",
	})
}

// initInfluxOptions sets options for direct InfluxDB measurement exporter.
func initInfluxOptions(
	c *cli.Cmd,
	influxEnabled **bool,
	influxEndpoint **string,
	influxDBName **string,
	influxUser **string,
	influxPassword **string,
) {
	*influxEnabled = c.Bool(cli.BoolOpt{
		Name:   "influx-enabled",
		Desc:   "Enables InfluxDB adapter and reporting",
		EnvVar: "COREMON_DB_INFLUX_ENABLED",
		Value:  true,
	})

	*influxEndpoint = c.String(cli.StringOpt{
		Name:   "influx-endpoint",
		Desc:   "Specify InfluxDB endpoint.",
		EnvVar: "COREMON_DB_INFLUX_ENDPOINT",
		Value:  "https://influx.docker.direct",
	})

	*influxDBName = c.String(cli.StringOpt{
		Name:   "influx-db-name",
		Desc:   "Specify InfluxDB database name.",
		EnvVar: "COREMON_DB_INFLUX_DBNAME",
		Value:  "telegraf",
	})

	*influxUser = c.String(cli.StringOpt{
		Name:   "influx-user",
		Desc:   "Specify InfluxDB user name.",
		EnvVar: "COREMON_DB_INFLUX_USERNAME",
		Value:  "coremon_user",
	})

	*influxPassword = c.String(cli.StringOpt{
		Name:   "influx-password",
		Desc:   "Specify InfluxDB database password.",
		EnvVar: "COREMON_DB_INFLUX_PASSWORD",
		Value:  "",
	})
}
