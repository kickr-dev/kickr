package generate_test

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"
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
