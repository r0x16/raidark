// Package ids_test verifies the public UUIDv7 helper contract and its use as a
// UUIDv7 primary key in Raidark GORM models.
package ids_test

import (
	"database/sql/driver"
	"sync"
	"testing"

	"github.com/google/uuid"
	datastoredomain "github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/r0x16/Raidark/shared/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type uuidEntity struct {
	datastoredomain.BaseModel
	Name string
}

// TestNewV7_returnsCanonicalRFCUUIDv7 verifies canonical formatting, version
// bits, and variant bits because consumers depend on that interoperable shape.
func TestNewV7_returnsCanonicalRFCUUIDv7(t *testing.T) {
	generated, err := ids.NewV7()
	require.NoError(t, err)

	parsed := uuid.MustParse(generated)
	assert.Len(t, generated, 36)
	assert.Equal(t, uuid.Version(7), parsed.Version())
	assert.Equal(t, uuid.RFC4122, parsed.Variant())
}

// TestNewV7_timestampsAreNonDecreasing covers the logical ordering property of
// UUIDv7: consecutive calls must not move backward in the embedded timestamp.
func TestNewV7_timestampsAreNonDecreasing(t *testing.T) {
	first, err := ids.NewV7()
	require.NoError(t, err)
	second, err := ids.NewV7()
	require.NoError(t, err)

	firstParsed := uuid.MustParse(first)
	secondParsed := uuid.MustParse(second)

	assert.GreaterOrEqual(t, int64(secondParsed.Time()), int64(firstParsed.Time()))
}

// TestIsValidV7_acceptsGeneratedIDs verifies that the validator accepts IDs
// emitted by the package generator itself.
func TestIsValidV7_acceptsGeneratedIDs(t *testing.T) {
	generated, err := ids.NewV7()
	require.NoError(t, err)

	assert.True(t, ids.IsValidV7(generated))
}

// TestIsValidV7_rejectsInvalidInputs fixes the expected rejection cases: empty
// input, malformed input, UUIDv4, and UUIDv7 with a non-RFC variant.
func TestIsValidV7_rejectsInvalidInputs(t *testing.T) {
	tests := map[string]string{
		"empty":         "",
		"malformed":     "not-a-uuid",
		"uuid-v4":       uuid.NewString(),
		"wrong-variant": "018bd12c-58b0-7683-0a5b-8752d0e86651",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			assert.False(t, ids.IsValidV7(input))
		})
	}
}

// TestNewV7_concurrentGenerationIsUnique forces bulk generation from multiple
// goroutines to catch collisions or race-safety issues under -race.
func TestNewV7_concurrentGenerationIsUnique(t *testing.T) {
	const totalIDs = 100_000
	const workers = 50

	results := make(chan string, totalIDs)
	errs := make(chan error, totalIDs)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < totalIDs/workers; i++ {
				generated, err := ids.NewV7()
				if err != nil {
					errs <- err
					return
				}
				results <- generated
			}
		}()
	}

	wg.Wait()
	close(results)
	close(errs)

	require.Empty(t, errs)

	seen := make(map[string]struct{}, totalIDs)
	for generated := range results {
		require.True(t, ids.IsValidV7(generated))
		if _, exists := seen[generated]; exists {
			t.Fatalf("UUIDv7 collision detected: %s", generated)
		}
		seen[generated] = struct{}{}
	}

	assert.Len(t, seen, totalIDs)
}

// TestUUIDv7_sqlInterfaces covers the minimum read/write contracts used by
// database/sql and GORM when mapping the custom type.
func TestUUIDv7_sqlInterfaces(t *testing.T) {
	generated, err := ids.NewV7()
	require.NoError(t, err)

	var fromString ids.UUIDv7
	require.NoError(t, fromString.Scan(generated))
	assert.Equal(t, ids.UUIDv7(generated), fromString)

	var fromBytes ids.UUIDv7
	require.NoError(t, fromBytes.Scan([]byte(generated)))
	assert.Equal(t, ids.UUIDv7(generated), fromBytes)

	assert.Error(t, fromBytes.Scan(123))

	value, err := ids.UUIDv7(generated).Value()
	require.NoError(t, err)
	assert.Equal(t, driver.Value(generated), value)
}

// TestBaseModel_beforeCreateAutogeneratesUUIDv7 verifies Raidark's canonical
// GORM integration: datastore/domain.BaseModel fills an empty UUIDv7 primary key
// and preserves an explicit value when the consumer has already set one.
func TestBaseModel_beforeCreateAutogeneratesUUIDv7(t *testing.T) {
	database := openUUIDTestDatabase(t)

	record := uuidEntity{Name: "generated"}
	require.NoError(t, database.Create(&record).Error)
	require.NotEmpty(t, record.ID)
	require.True(t, ids.IsValidV7(string(record.ID)))

	var found uuidEntity
	require.NoError(t, database.First(&found, "id = ?", record.ID).Error)
	assert.Equal(t, record.ID, found.ID)

	explicit, err := ids.NewV7()
	require.NoError(t, err)
	withExplicitID := uuidEntity{
		BaseModel: datastoredomain.BaseModel{ID: ids.UUIDv7(explicit)},
		Name:      "explicit",
	}

	require.NoError(t, database.Create(&withExplicitID).Error)
	assert.Equal(t, ids.UUIDv7(explicit), withExplicitID.ID)
}

// openUUIDTestDatabase creates an isolated SQLite database for validating GORM
// hooks without depending on external services or product migrations.
func openUUIDTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&uuidEntity{}))

	return database
}
