package schemas

import (
	"embed"
	"io/fs"
)

//go:generate go run ./gen/main.go

//go:embed *
var embedded embed.FS

var _ fs.FS = (*embed.FS)(nil) // ensure interface is implemented

func FS() fs.FS {
	return embedded
}
