package store

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
)

type FileConfig struct {
	Dir string `yaml:"dir" json:"dir" mapstructure:"dir" env:"STORE_FILE_DIR"`
}

// FileStore File-based cache implementation, each key corresponds to a file
type FileStore struct {
	dir   string
	locks sync.Map // key -> *sync.RWMutex
}

// NewFileStore Creates a new file cache instance
func NewFileStore(config FileConfig) (*FileStore, error) {
	if config.Dir == "" {
		config.Dir = os.TempDir()
	}
	if err := os.MkdirAll(config.Dir, 0755); err != nil {
		return nil, err
	}
	return &FileStore{dir: config.Dir}, nil
}

func (fs *FileStore) filePath(key string) string {
	hash := md5.Sum([]byte(key))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(fs.dir, filename)
}

// getLock Acquires the lock for the given key
func (fs *FileStore) getLock(key string) *sync.RWMutex {
	val, _ := fs.locks.LoadOrStore(key, &sync.RWMutex{})
	return val.(*sync.RWMutex)
}

// Has Checks if the key exists
func (fs *FileStore) Has(key string) bool {
	lock := fs.getLock(key)
	lock.RLock()
	defer lock.RUnlock()
	_, err := os.Stat(fs.filePath(key))
	return err == nil
}

// Set Writes the value corresponding to the key into a file (atomic write)
func (fs *FileStore) Set(key string, value []byte) error {
	lock := fs.getLock(key)
	lock.Lock()
	defer lock.Unlock()
	return os.WriteFile(fs.filePath(key), value, 0644)
}

// Get Reads the value corresponding to the key from the file
func (fs *FileStore) Get(key string) ([]byte, error) {
	lock := fs.getLock(key)
	lock.RLock()
	defer lock.RUnlock()
	return os.ReadFile(fs.filePath(key))
}

// Delete Deletes the file corresponding to the key
func (fs *FileStore) Delete(key string) error {
	lock := fs.getLock(key)
	lock.Lock()
	defer lock.Unlock()
	path := fs.filePath(key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Consider the delete operation successful if the file doesn't exist
	}
	return os.Remove(path)
}
