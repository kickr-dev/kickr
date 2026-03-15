package generate_test

import (
	"os"
	"testing"

	log "charm.land/log/v2"
	engine "github.com/kickr-dev/engine/pkg"
)

func TestMain(m *testing.M) {
	engine.Configure(
		engine.WithForce(true),
		engine.WithLogger(log.NewWithOptions(os.Stderr, log.Options{
			CallerFormatter: log.ShortCallerFormatter,
			Level:           log.WarnLevel,
			ReportCaller:    true,
		})))
	m.Run()
}
