package schemas

import (
	"embed"
	"io/fs"
)

//go:generate go run github.com/mikefarah/yq/v4@v4.54.1 --output-format json --indent 0 --split-exp "\"kickr.v1.schema.json\"" kickr.v1.schema.yml

//go:embed *.json
var embedded embed.FS

var _ fs.FS = (*embed.FS)(nil) // ensure interface is implemented

// FS returns the embedded JSON schemas.
func FS() fs.FS {
	return embedded
}
