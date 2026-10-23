package schemas

import (
	"embed"
	"io/fs"
)

//go:generate go run ./gen/main.go

//go:embed *.json
var embedded embed.FS

var _ fs.FS = (*embed.FS)(nil) // ensure interface is implemented

// FS returns the embedded JSON schemas.
func FS() fs.FS {
	return embedded
}
