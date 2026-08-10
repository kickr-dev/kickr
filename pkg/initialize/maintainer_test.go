package initialize_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/initialize"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestMaintainer(t *testing.T) {
	ctx := t.Context()

	t.Run("success_full", func(t *testing.T) {
		// Arrange
		reader := NewPacedReader("John Doe"+defaultSubmit+"john@example.com"+defaultSubmit+"https://john.dev"+defaultSubmit, 2*time.Millisecond)
		expected := kickr.Kickr{Maintainers: []*kickr.Maintainer{{Name: "John Doe", Email: "john@example.com", URL: "https://john.dev"}}}

		// Act
		config, err := engine.Initialize(ctx,
			engine.WithFormGroups(initialize.Maintainer),
			engine.WithTeaOptions[kickr.Kickr](tea.WithInput(reader)))

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_optional_fields_empty", func(t *testing.T) {
		// Arrange
		reader := NewPacedReader("John Doe"+defaultSubmit+defaultSubmit+defaultSubmit, 2*time.Millisecond)
		expected := kickr.Kickr{Maintainers: []*kickr.Maintainer{{Name: "John Doe"}}}

		// Act
		config, err := engine.Initialize(ctx,
			engine.WithFormGroups(initialize.Maintainer),
			engine.WithTeaOptions[kickr.Kickr](tea.WithInput(reader)))

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}
