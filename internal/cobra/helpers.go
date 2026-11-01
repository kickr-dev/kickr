package cobra

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "charm.land/log/v2"
	"github.com/spf13/cobra"
	engine "gitlab.com/kickr-dev/engine/pkg"
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

func setupForce(cmd *cobra.Command, force *bool) error {
	if cmd.Flags().Changed(flagForce) {
		return nil
	}
	env := getenv(envPrefix + flagForce)
	if env == "" {
		return nil
	}
	ff, err := strconv.ParseBool(env)
	if err != nil {
		return fmt.Errorf(`invalid argument %q for "--%s" flag: %w`, env, flagForce, err)
	}
	*force = ff
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

func httpClient() *http.Client {
	return &http.Client{Transport: engine.NewLoggerURL(http.DefaultTransport)}
}

func usage(parent, child *cobra.Command, err error) {
	for _, help := range []string{"required flag(s)", "unknown command", "unknown flag", "unknown shorthand flag"} {
		if !strings.Contains(err.Error(), help) {
			continue
		}

		if child != nil {
			child.Println(child.UsageString())
			break
		}
		parent.Println(parent.UsageString())
		break
	}
}
