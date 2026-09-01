package writer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/antigravity/kogen/internal/core/domain"
)

type FSWriterAdapter struct{}

func NewFSWriterAdapter() *FSWriterAdapter {
	return &FSWriterAdapter{}
}

func (w *FSWriterAdapter) WriteFiles(ctx context.Context, baseDir string, files []domain.GeneratedFile, force bool, dryRun bool) error {
	for _, f := range files {
		fullPath := filepath.Join(baseDir, f.Path)

		if dryRun {
			fmt.Printf("[DRY-RUN] Would create: %s (%d bytes)\n", fullPath, len(f.Content))
			continue
		}

		// Ensure directory exists
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Check if file exists
		if _, err := os.Stat(fullPath); err == nil && !force {
			fmt.Printf("[SKIP] File exists: %s (use --force to overwrite)\n", fullPath)
			continue
		}

		perm := os.FileMode(0644)
		if f.IsExecutable {
			perm = os.FileMode(0755)
		}

		if err := os.WriteFile(fullPath, []byte(f.Content), perm); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		fmt.Printf("[CREATED] %s\n", fullPath)
	}

	return nil
}
