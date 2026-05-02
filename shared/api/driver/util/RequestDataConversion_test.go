// Package util_test verifies request parsing helpers used by API controllers.
package util_test

import (
	"testing"
	"time"

	apiutil "github.com/r0x16/Raidark/shared/api/driver/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *time.Time
		wantError string
	}{
		{name: "empty input", input: "", want: nil},
		{name: "valid date", input: "02-05-2026", want: mustDate(t, "2026-05-02")},
		{name: "invalid layout", input: "2026-05-02", wantError: "invalid date format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apiutil.ParseDate(tt.input)

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.True(t, tt.want.Equal(*got))
		})
	}
}

func TestParsePage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty defaults to first page", input: "", want: 1},
		{name: "invalid defaults to first page", input: "abc", want: 1},
		{name: "valid integer", input: "7", want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, apiutil.ParsePage(tt.input))
		})
	}
}

func TestParsePageSize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int
		wantError string
	}{
		{name: "empty defaults to ten", input: "", want: 10},
		{name: "invalid defaults to ten", input: "large", want: 10},
		{name: "valid integer", input: "25", want: 25},
		{name: "capped at one hundred", input: "101", want: 100, wantError: "size cannot be bigger than 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apiutil.ParsePageSize(tt.input)

			assert.Equal(t, tt.want, got)
			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParseUintID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint
		wantError string
	}{
		{name: "empty input", input: "", wantError: "id cannot be empty"},
		{name: "negative value", input: "-1", wantError: "invalid id format"},
		{name: "non numeric", input: "abc", wantError: "invalid id format"},
		{name: "valid id", input: "42", want: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apiutil.ParseUintID(tt.input)

			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				assert.Zero(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func mustDate(t *testing.T, value string) *time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	require.NoError(t, err)
	return &parsed
}
