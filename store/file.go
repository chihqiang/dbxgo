package store

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
)

type FileConfig struct {
	Dir string `yaml:"dir"`
}

// FileStore 文件缓存实现，每个 key 对应一个文件
type FileStore struct {
	dir   string
	locks sync.Map // key -> *sync.RWMutex
}

// NewFileStore 创建文件缓存实例
func NewFileStore(config FileConfig) (*FileStore, error) {
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

// getLock 获取 key 的锁
func (fs *FileStore) getLock(key string) *sync.RWMutex {
	val, _ := fs.locks.LoadOrStore(key, &sync.RWMutex{})
	return val.(*sync.RWMutex)
}

// Has 判断 key 是否存在
func (fs *FileStore) Has(key string) bool {
	lock := fs.getLock(key)
	lock.RLock()
	defer lock.RUnlock()
	_, err := os.Stat(fs.filePath(key))
	return err == nil
}

// Set 写入 key 对应的文件（原子写入）
func (fs *FileStore) Set(key string, value []byte) error {
	lock := fs.getLock(key)
	lock.Lock()
	defer lock.Unlock()
	return os.WriteFile(fs.filePath(key), value, 0644)
}

// Get 读取 key 对应的文件
func (fs *FileStore) Get(key string) ([]byte, error) {
	lock := fs.getLock(key)
	lock.RLock()
	defer lock.RUnlock()
	return os.ReadFile(fs.filePath(key))
}
