package cobra

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	CallerFormatter: log.ShortCallerFormatter,
	ReportCaller:    true,
})

func setupLogger(logFormat, logLevel string) error {
	styles := log.DefaultStyles()
	switch logFormat {
	case "text":
		logger.SetFormatter(log.TextFormatter)
		for _, level := range []log.Level{log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.FatalLevel} {
			styles.Levels[level] = styles.Levels[level].MaxWidth(len(level.String()))
		}
		logger.SetStyles(styles)
	case "json":
		logger.SetFormatter(log.JSONFormatter)
	default:
		return fmt.Errorf(`invalid argument %q for "--%s" flag`, logFormat, flagLogFormat)
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf(`invalid argument %q for "--%s" flag`, logLevel, flagLogLevel)
	}
	logger.SetLevel(level)

	return nil
}

func coalesce(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func getenv(flag string) string {
	key := strings.ToUpper(strings.ReplaceAll(flag, "-", "_"))
	return os.Getenv(key)
}
