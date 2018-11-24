package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	golog "log"
	"os"
	"time"
)

// Minimum log level to accept for output. Any log statements with a lower level will not be printed. Can be:
//  * trace
//  * debug
//  * info
//  * warn
//  * error
var logLevel string

// Whether to include caller information for each log entry. This has significant performance overhead and thus should
// only be used in debugging sessions or local development.
var caller bool

// Root command serving as the root commands tree
var rootCmd = &cobra.Command{
	Use:     "gcp",
	Version: "1.0.0-alpha.1", // TODO: externalize agent version
	Short:   "Google Cloud Platform resources",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
	},
}

// Initializes the main package with global flags
func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().BoolVarP(&caller, "caller", "c", false, "Include caller information in log output")
}

// Initializes the logging framework
func initLogging() {
	log.SetOutput(os.Stdout)
	log.SetReportCaller(caller)

	// apply log level
	switch logLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		golog.Fatalf("invalid loglevel provided: %s\n", logLevel)
	}

	// apply appropriate log formatter, according to the logFormat flag
	log.SetFormatter(&log.JSONFormatter{
		DisableTimestamp: false,
		PrettyPrint:      false,
		TimestampFormat:  time.RFC3339,
	})

	// redirect Golang standard log package output to logrus
	golog.SetFlags(0)
	golog.SetOutput(log.StandardLogger().Writer())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Execution error")
	}
}
