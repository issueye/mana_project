package store

import (
	"context"
	"fmt"
	"os"
)

// Backup writes a consistent snapshot of the database to dest with
// `VACUUM INTO`, which works while the app keeps using the original. The
// destination must not exist yet.
func (s *Store) Backup(ctx context.Context, dest string) error {
	if dest == "" {
		return fmt.Errorf("backup: destination is empty")
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("backup: %s already exists", dest)
	}
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, dest); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	return nil
}
