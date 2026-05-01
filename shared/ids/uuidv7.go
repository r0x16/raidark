package ids

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// UUIDv7 is a string-based GORM type for UUID v7 primary keys.
type UUIDv7 string

// NewV7 generates a new UUID v7 string in canonical form (8-4-4-4-12 hex).
func NewV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("ids: uuid v7 generation failed: %w", err)
	}
	return id.String(), nil
}

// IsValidV7 returns true if s is a valid UUID v7 per RFC 9562:
// correct canonical form, version 7, and RFC 4122 variant bits (10xxxxxx).
func IsValidV7(s string) bool {
	id, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return id.Version() == 7 && id.Variant() == uuid.RFC4122
}

// Scan implements sql.Scanner for reading from a database column.
func (u *UUIDv7) Scan(value any) error {
	switch v := value.(type) {
	case string:
		*u = UUIDv7(v)
	case []byte:
		*u = UUIDv7(v)
	default:
		return fmt.Errorf("ids: cannot scan %T into UUIDv7", value)
	}
	return nil
}

// Value implements driver.Valuer for writing to a database column.
func (u UUIDv7) Value() (driver.Value, error) {
	return string(u), nil
}
