package kickr

import "github.com/goccy/go-yaml"

// EncodeOpts returns the options related to YAML encoding with goccy/go-yaml.
func EncodeOpts() []yaml.EncodeOption {
	return []yaml.EncodeOption{
		yaml.Indent(2),
		yaml.IndentSequence(true),
		yaml.WithComment(yaml.CommentMap{
			"$": []*yaml.Comment{
				yaml.HeadComment(
					" Kickr configuration file (https://github.com/kickr-dev/kickr)",
					" yaml-language-server: $schema=https://raw.githubusercontent.com/kickr-dev/kickr/beta/.schemas/kickr.schema.json",
				),
			},
		}),
	}
}
