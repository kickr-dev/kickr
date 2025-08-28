package schemas

import "embed"

//go:generate go run ./gen/main.go

//go:embed *
var fs embed.FS

// ReadFile reads the input name from .schemas embedded fs.
func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(name)
}
