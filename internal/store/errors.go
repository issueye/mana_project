package store

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("store: not found")

// ErrConflict is returned when a unique constraint fails.
var ErrConflict = errors.New("store: conflict")

func mapError(op string, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "constraint failed") {
		return fmt.Errorf("%w: %s", ErrConflict, msg)
	}
	return fmt.Errorf("%s: %w", op, err)
}
