package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	golog "log"
	"os"
	"time"
)

// Workspace to place all build request workspaces in
var workspacePath string

// Log output format; can be "auto", "json", "plain" or "pretty":
//  * "auto": if a TTY is attached, acts the same as "pretty"; otherwise uses "json"
//  * "json": each log entry will be a JSON object containing all available information such as msg, timestamp, etc
//  * "plain": human-friendly output (unlike JSON) but without ANSI colors
//  * "pretty": human-friendly output with ANSI colors
var logFormat string

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
	Use:     "agent",
	Version: "1.0.0-alpha.1", // TODO: externalize agent version
	Short:   "Gitzup agent executes pipelines",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
	},
}

// Initializes the main package with global flags
func init() {
	rootCmd.PersistentFlags().StringVarP(&workspacePath, "workspace", "w", ".", "Workspace location")
	rootCmd.PersistentFlags().StringVar(&logFormat, "logformat", "auto", "Log output format (auto, json, plain, pretty)")
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
	switch logFormat {
	case "auto":
		// no-op, auto-detected by logrus
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: false,
			PrettyPrint:      false,
			TimestampFormat:  time.RFC3339,
		})
	case "plain":
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp:       false,
			DisableColors:          true,
			DisableLevelTruncation: true,
			DisableSorting:         false,
			ForceColors:            false,
			FullTimestamp:          true,
			TimestampFormat:        time.RFC3339,
		})
	case "pretty":
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp:       false,
			DisableColors:          false,
			DisableLevelTruncation: false,
			DisableSorting:         false,
			ForceColors:            true,
			FullTimestamp:          true,
			TimestampFormat:        time.RFC3339,
		})
	default:
		golog.Fatalf("invalid logformat provided: %s\n", logFormat)
	}

	// redirect Golang standard log package output to logrus
	golog.SetFlags(0)
	golog.SetOutput(log.StandardLogger().Writer())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Execution error")
	}
}
