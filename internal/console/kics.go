package console

import (
	"context"
	"io"
	"os"
	"time"

	consoleHelpers "github.com/Checkmarx/kics/internal/console/helpers"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	scanID   = "console"
	timeMult = 2
)

var (
	ctx = context.Background()

	verbose  bool
	logFile  bool
	logLevel string

	rootCmd = &cobra.Command{
		Use:   "kics",
		Short: "Keeping Infrastructure as Code Secure",
	}
)

func initialize() error {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(generateIDCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.PersistentFlags().BoolVarP(&logFile, "log-file", "l", false, "writes log messages to info.log")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "INFO", "determines log level (TRACE,DEBUG,INFO,WARN,ERROR,FATAL)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "increase verbosity")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}

	initScanCmd()
	if insertScanCmd() {
		log.Warn().Msg("DEPRECATION WARNING: adding 'scan' sub-command as argument" +
			" for future versions use 'kics scan'")
		os.Args = append([]string{os.Args[0], "scan"}, os.Args[1:]...)
	}
	return nil
}

func setLogLevel() {
	switch logLevel {
	case "TRACE":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		log.Warn().Msg("invalid log level, setting default INFO level")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func setupLogs() error {
	consoleLogger := zerolog.ConsoleWriter{Out: io.Discard}
	fileLogger := zerolog.ConsoleWriter{Out: io.Discard}

	setLogLevel()

	if verbose {
		log.Debug().Msg("verbose mode, redirecting logs to stdout")
		consoleLogger = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	if logFile {
		log.Debug().Msg("creating info.log file")
		file, err := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		fileLogger = consoleHelpers.CustomConsoleWriter(&zerolog.ConsoleWriter{Out: file, NoColor: true})
	}

	mw := io.MultiWriter(consoleLogger, fileLogger)
	log.Logger = log.Output(mw)
	return nil
}

func insertScanCmd() bool {
	if len(os.Args) > 1 && os.Args[1][0] == '-' {
		if os.Args[1][1] != '-' {
			flag := os.Args[1][1:]
			return scanCmd.Flags().ShorthandLookup(flag) != nil
		}
		flag := os.Args[1][2:]
		return scanCmd.Flag(flag) != nil
	}
	return false
}

// Execute starts kics execution
func Execute() {
	defer sentry.Flush(timeMult * time.Second)
	if err := initialize(); err != nil {
		sentry.CaptureException(err)
		log.Err(err).Msg("failed to run application")
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		sentry.CaptureException(err)
		log.Err(err).Msg("failed to run application")
	}

	if err := setupLogs(); err != nil {
		sentry.CaptureException(err)
		log.Err(err).Msg("failed to initialize logs")
	}
}
