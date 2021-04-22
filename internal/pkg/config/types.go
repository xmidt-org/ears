package config

// Config interface for uber/fx
type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}
