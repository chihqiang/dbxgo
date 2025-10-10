package store

import (
	"fmt"
)

type StoreType string

const (
	// FileStoreType Type for File Store
	FileStoreType StoreType = "file"
	// RedisStoreType Type for Redis Store
	RedisStoreType StoreType = "redis"
)

var (
	// stores holds the registered store creators for different types
	stores = map[StoreType]func(cfg Config) (IStore, error){
		// For file store, create a new file store instance
		FileStoreType: func(cfg Config) (IStore, error) {
			return NewFileStore(cfg.File)
		},
		// For Redis store, create a new Redis store instance
		RedisStoreType: func(cfg Config) (IStore, error) {
			return NewRedisStore(cfg.Redis)
		},
	}
)

// Register registers a custom store creator function for a given store type
func Register(storeType StoreType, fn func(Config) (IStore, error)) {
	stores[storeType] = fn
}

// Config Storage configuration structure
// Used to configure the storage type and related configurations (e.g., file or Redis)
type Config struct {
	// Type Storage type, e.g., "file"
	Type  StoreType   `yaml:"type" json:"type" mapstructure:"type" env:"STORE_TYPE,required"`
	File  FileConfig  `yaml:"file" json:"file" mapstructure:"file"`
	Redis RedisConfig `yaml:"redis" json:"redis" mapstructure:"redis"`
}

// IStore Interface for storage operations
// All concrete store types must implement this interface
type IStore interface {
	// Set stores the value associated with the key
	Set(key string, value []byte) error
	// Get retrieves the value associated with the key
	Get(key string) ([]byte, error)
	// Has checks if the key exists
	Has(key string) bool
	// Delete removes the key-value pair
	Delete(key string) error
}

// NewStore Creates a new store instance based on the provided configuration
// cfg: The configuration for the store
// Returns the store instance and any error encountered during creation
func NewStore(cfg Config) (IStore, error) {
	// Look up the store creation function based on the store type
	creator, exists := stores[cfg.Type]
	if !exists {
		return nil, fmt.Errorf("store type %s is not registered", cfg.Type)
	}
	// Call the creation function to produce the store instance
	return creator(cfg)
}
