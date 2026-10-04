package generate

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/generator"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// HugoImage is the container image repository of Hugo, whose version digest is pinned by ParserHugo.
const HugoImage = "ghcr.io/gohugoio/hugo"

// ParserHugo detects Hugo sites in repository modules and sets their language accordingly.
//
// When a module defines a hugo version (max, else min), its image is fetched and pinned by digest.
func ParserHugo(httpClient *http.Client) func(ctx context.Context, destdir string, repo *types.Repository) error {
	if httpClient == nil {
		httpClient = http.DefaultClient //nolint:revive
	}
	return func(ctx context.Context, destdir string, repo *types.Repository) error {
		errs := make([]error, 0, len(repo.Modules))
		for i, module := range repo.Modules {
			hugo, err := parser.ReadHugo(filepath.Join(destdir, module.Dir()))
			if err != nil {
				if !errors.Is(err, parser.ErrNoHugo) {
					errs = append(errs, fmt.Errorf("read hugo in '%s': %w", module.Dir(), err))
				}
				continue
			}
			engine.GetLogger().Infof("hugo detected in '%s', theme or hugo files are present", module.Dir())

			version := hugo.Module().HugoVersion
			if v := cmp.Or(version.Max, version.Min); v != "" {
				image, err := generator.FetchContainerImage(ctx, httpClient, HugoImage, "v"+strings.TrimPrefix(v, "v"))
				if err != nil {
					engine.GetLogger().Warnf("failed to fetch hugo image in '%s', using default one: %v", module.Dir(), err)
				}
				hugo.ContainerImage = image
			}
			repo.Modules[i].SetLanguage(types.LanguageHugo, hugo)
		}
		return errors.Join(errs...) // already wrapped
	}
}

var _ engine.Parser[types.Repository] = ParserHugo(nil) // ensure interface is implemented
