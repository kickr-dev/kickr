package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// GeneratorLicense generates the license file for the project.
func GeneratorLicense(httpClient *http.Client) func(ctx context.Context, destdir string, repo types.Repository) error {
	if httpClient == nil {
		httpClient = http.DefaultClient //nolint:revive
	}
	return func(ctx context.Context, destdir string, repo types.Repository) error {
		dest := filepath.Join(destdir, "LICENSE")
		if repo.Config.License == "" {
			engine.GetLogger().Infof("skipping license generation, configuration doesn't have 'license' key")
			if err := os.Remove(dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("remove '%s': %w", "LICENSE", err)
			}
			return nil
		}

		client, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"),
			gitlab.WithBaseURL("https://gitlab.com/api/v4"),
			gitlab.WithHTTPClient(httpClient),
			gitlab.WithoutRetries(),
			gitlab.WithRequestOptions(gitlab.WithContext(ctx)))
		if err != nil {
			// should never happen since it's gitlab.ClientOptionFunc that are throwing errors
			// and currently WithBaseURL with fixed URL
			// and WithoutRetries won't throw errors
			// but in any case err must be handled in case it evolves or other options are added
			engine.GetLogger().Warnf("failed to initialize GitLab client, skipping license generation: %v", err)
			return nil
		}

		if !engine.Forced() && files.Exists(dest) {
			engine.GetLogger().Infof("not generating '%s' since it already exists", "LICENSE")
			return nil
		}
		engine.GetLogger().Infof("license detected, configuration has 'license' key")

		opts := gitlab.GetLicenseTemplateOptions{
			Fullname: func() *string {
				var zero string
				if len(repo.Config.Maintainers) == 0 {
					return &zero
				}
				if repo.Config.Maintainers[0] == nil {
					return &zero
				}
				return &repo.Config.Maintainers[0].Name
			}(),
			Project: &repo.VCS.ProjectName,
		}
		template, _, err := client.LicenseTemplates.GetLicenseTemplate(repo.Config.License, &opts)
		if err != nil {
			return fmt.Errorf("get license template '%s': %w", repo.Config.License, err)
		}
		if err := os.WriteFile(dest, []byte(template.Content), files.RwRR); err != nil { //nolint:gosec
			return fmt.Errorf("write license file: %w", err)
		}
		return nil
	}
}

var _ engine.Generator[types.Repository] = GeneratorLicense(nil) // ensure interface is implemented
