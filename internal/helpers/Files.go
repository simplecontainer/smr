package helpers

import (
	"context"
	"fmt"
	"os"
	"time"
)

func WaitForFileToAppear(ctx context.Context, path string, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled or timeout: %w", ctx.Err())
		case <-ticker.C:
			if _, err := os.Stat(path); err == nil {
				return nil // file found
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("error checking file: %w", err)
			}
		}
	}
}
