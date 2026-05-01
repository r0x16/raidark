package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_smoke(t *testing.T) {
	contents := Read(t, "sample.txt")

	assert.Equal(t, "raidark fixture\n", string(contents))
}

func TestFS_smoke(t *testing.T) {
	filesystem := FS(t)

	contents, err := filesystem.Open("sample.txt")
	require.NoError(t, err)
	require.NoError(t, contents.Close())
}
