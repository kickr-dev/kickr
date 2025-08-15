package schemas

import "embed"

const (
	Chart = "chart.schema.json"
	Kickr = "kickr.schema.json"
)

//go:embed *.json
var fs embed.FS

// ReadFile reads the input name from .schemas embedded fs.
func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(name)
}
