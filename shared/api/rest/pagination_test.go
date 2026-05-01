// Package rest_test verifies Raidark's public REST pagination contract from
// the point of view of services and HTTP clients.
package rest_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/r0x16/Raidark/shared/api/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCursor struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
}

// TestPage_jsonShape verifies the generic page envelope remains stable for
// consumers that parse list responses with shared clients.
func TestPage_jsonShape(t *testing.T) {
	page := rest.Page[string]{
		Items: []string{"alpha", "beta"},
		Pagination: rest.PageMeta{
			NextCursor: "opaque-cursor",
			Limit:      25,
		},
	}

	data, err := json.Marshal(page)
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"items": ["alpha", "beta"],
		"pagination": {
			"next_cursor": "opaque-cursor",
			"limit": 25
		}
	}`, string(data))
}

// TestCursor_roundTrip verifies cursor encoding can carry compound keyset
// state through an HTTP query parameter and back into typed Go values.
func TestCursor_roundTrip(t *testing.T) {
	input := testCursor{
		CreatedAt: "2026-05-01T12:00:00Z",
		ID:        "018f46c0-0000-7000-8000-000000000001",
	}

	cursor, err := rest.EncodeCursor(input)
	require.NoError(t, err)

	var output testCursor
	require.NoError(t, rest.DecodeCursor(cursor, &output))

	assert.Equal(t, input, output)
	assert.NotContains(t, cursor, "=")
	assert.NotContains(t, cursor, "+")
	assert.NotContains(t, cursor, "/")
}

// TestCursor_encodeRejectsUnmarshalableValues documents the failure mode for
// invalid cursor state before any opaque token is returned to a client.
func TestCursor_encodeRejectsUnmarshalableValues(t *testing.T) {
	cursor, err := rest.EncodeCursor(map[string]any{
		"callback": func() {},
	})

	assert.Empty(t, cursor)
	assert.Error(t, err)
}

// TestCursor_tamperingInvalidatesCursor covers the malformed-cursor path a
// client hits after changing bytes in the opaque token.
func TestCursor_tamperingInvalidatesCursor(t *testing.T) {
	cursor, err := rest.EncodeCursor(testCursor{ID: "018f46c0-0000-7000-8000-000000000001"})
	require.NoError(t, err)

	tampered := cursor[:len(cursor)-1] + "!"

	var output testCursor
	assert.Error(t, rest.DecodeCursor(tampered, &output))
}

// TestCursor_invalidPayloadIsRejected verifies decoding fails even when base64
// itself is valid but the decoded bytes are not a cursor JSON payload.
func TestCursor_invalidPayloadIsRejected(t *testing.T) {
	cursor := base64.RawURLEncoding.EncodeToString([]byte("not-json"))

	var output testCursor
	assert.Error(t, rest.DecodeCursor(cursor, &output))
}

// TestClampLimit_defaultsAndClamps verifies the default limit for missing query
// input and the upper bound used to protect list endpoints.
func TestClampLimit_defaultsAndClamps(t *testing.T) {
	tests := map[string]struct {
		input int
		want  int
	}{
		"missing query uses default":  {input: 0, want: rest.DefaultLimit},
		"negative uses default":       {input: -10, want: rest.DefaultLimit},
		"within range is preserved":   {input: 50, want: 50},
		"above max clamps to maximum": {input: rest.MaxLimit + 1, want: rest.MaxLimit},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, rest.ClampLimit(test.input))
		})
	}
}
