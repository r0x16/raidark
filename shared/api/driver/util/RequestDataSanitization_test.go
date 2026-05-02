// Package util_test verifies string sanitization helpers used before request
// data reaches application services.
package util_test

import (
	"testing"

	apiutil "github.com/r0x16/Raidark/shared/api/driver/util"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty input", input: "", want: ""},
		{name: "trims ascii whitespace", input: "  Raidark  ", want: "Raidark"},
		{name: "escapes html", input: `<script>alert("x")</script>`, want: `&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt;`},
		{name: "preserves unicode text", input: "  Ñandú  ", want: "Ñandú"},
		{name: "keeps internal control characters", input: "line\u0000break", want: "line\u0000break"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, apiutil.SanitizeString(tt.input))
		})
	}
}
