// Package domain_test verifies the public storage key helper contract used by
// storage drivers and service code before writing object bytes.
package domain_test

import (
	"strings"
	"testing"

	"github.com/r0x16/Raidark/shared/ids"
	"github.com/r0x16/Raidark/shared/storage/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildKey_returnsValidStorageKey verifies that generated keys pass the
// same validation contract enforced for caller-provided keys.
func TestBuildKey_returnsValidStorageKey(t *testing.T) {
	key, err := domain.BuildKey("accounts", "avatar", "png")

	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(key, ".png"))
	assert.NoError(t, domain.ValidateKey(key))
}

// TestValidateKey_acceptsCanonicalKeyWithOptionalExtension fixes the accepted
// storage key shape with and without a filename extension.
func TestValidateKey_acceptsCanonicalKeyWithOptionalExtension(t *testing.T) {
	id, err := ids.NewV7()
	require.NoError(t, err)

	assert.NoError(t, domain.ValidateKey("accounts/avatar/2026/05/"+id+".webp"))
	assert.NoError(t, domain.ValidateKey("accounts/avatar/2026/05/"+id))
}

// TestValidateKey_rejectsMalformedKeys covers traversal and convention errors
// before callers can pass unsafe or unaddressable object keys to a driver.
func TestValidateKey_rejectsMalformedKeys(t *testing.T) {
	id, err := ids.NewV7()
	require.NoError(t, err)

	tests := map[string]string{
		"empty":             "",
		"absolute":          "/accounts/avatar/2026/05/" + id + ".png",
		"parent-segment":    "accounts/../2026/05/" + id + ".png",
		"current-segment":   "accounts/./2026/05/" + id + ".png",
		"missing-namespace": "/avatar/2026/05/" + id + ".png",
		"missing-usage":     "accounts//2026/05/" + id + ".png",
		"invalid-year":      "accounts/avatar/26/05/" + id + ".png",
		"invalid-month":     "accounts/avatar/2026/13/" + id + ".png",
		"invalid-uuid":      "accounts/avatar/2026/05/not-a-uuid.png",
	}

	for name, key := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, domain.ValidateKey(key))
		})
	}
}

// TestValidateKey_allowsLiteralDotsInsideSegmentNames prevents regressions to
// substring-based traversal checks that reject safe names such as "a..b".
func TestValidateKey_allowsLiteralDotsInsideSegmentNames(t *testing.T) {
	id, err := ids.NewV7()
	require.NoError(t, err)

	key := "accounts/a..b/2026/05/" + id + ".png"

	assert.NoError(t, domain.ValidateKey(key))
}
