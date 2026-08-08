package schemas

import (
	"embed"
	"io/fs"
)

//go:generate go run ./gen/main.go

//go:embed *
var embedded embed.FS

func FS() fs.FS {
	return embedded
}
