package domain

type EnvProvider interface {
	GetString(key, defaultValue string) string
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetFloat(key string, defaultValue float64) float64
	GetSlice(key string, defaultValue []string) []string
	GetSliceWithSeparator(key, separator string, defaultValue []string) []string
	IsSet(key string) bool
	MustGet(key string) string
}
