package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type FileStore struct {
	folder      string
	knownValues sync.Map
}

// Has from a NoopStore will always return false, nil.
func (f *FileStore) Has(key string) (bool, error) {
	return false, nil
}

// Get will return os.ErrNotExist.
func (f *FileStore) Get(_ io.Writer, key string) error {
	return os.ErrNotExist
}

// Put writes content into file.
func (f *FileStore) Put(key string, content io.Reader) error {
	hashValue := calcHash(key)
	cachePath := path.Join(f.folder, hashValue)
	data, _ := ioutil.ReadAll(content)
	err := ioutil.WriteFile(cachePath, data, 0644)

	// Make sure, that the RAM-cache only holds values we were able to write.
	// This is a decision to prevent a false impression of the cache: If the
	// write fails, the cache isn't working correctly, which should be fixed by
	// the user of this cache.
	if err == nil {
		Debug.Printf("Cache wrote content into '%s'", cachePath)
		// f.knownValues.Store(hashValue, content)
	}

	return err
}

type Cache struct {
	folder      string
	hash        hash.Hash
	knownValues map[string][]byte
	mutex       sync.RWMutex
}

func CreateCache(path string) (*Cache, error) {
	// fileInfos, err := ioutil.ReadDir(path)
	// if err != nil {
	// 	Error.Printf("Error opening cache folder '%s':\n", path)
	// 	return nil, err
	// }

	// values := make(map[string][]byte, 0)

	// Go through every file an save its name in the map. The content of the file
	// is loaded when needed. This makes sure that we don't have to read
	// the directory content each time the user wants data that's not yet loaded.
	// for _, info := range fileInfos {
	// 	if !info.IsDir() {
	// 		values[info.Name()] = nil
	// 	}
	// }

	hash := sha256.New()

	cache := &Cache{
		folder: path,
		hash:   hash,
		// knownValues: values,
		mutex: sync.RWMutex{},
	}

	return cache, nil
}

// CacheKey generate cache key hash
func (c *Cache) CacheKey(key string) string {
	return calcHash(key)
}
func (c *Cache) clear() error {
	d, err := os.Open(c.folder)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(c.folder, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) remove(key string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	hashValue := calcHash(key)
	// _, ok := c.knownValues[hashValue]
	filepath := path.Join(c.folder, hashValue)
	if _, err := os.Stat(filepath); err == nil {
		return os.Remove(filepath)
	}
	return nil
}
func (c *Cache) has(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	hashValue := calcHash(key)
	// _, ok := c.knownValues[hashValue]

	if _, err := os.Stat(path.Join(c.folder, hashValue)); err == nil {
		return hashValue, true
	}
	return hashValue, false
	// return ok
}

// Get item from cache
func (c *Cache) get(key string) (string, []byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	hashValue := calcHash(key)

	// Try to get content. Error if not found.
	Debug.Printf("Try to get key '%s'", key)

	// c.mutex.Lock()
	// content, ok := c.knownValues[hashValue]
	// c.mutex.Unlock()
	// if !ok {
	// 	Debug.Printf("Cache doesn't know key '%s'", hashValue)
	// 	return nil, errors.New(fmt.Sprintf("Key '%s' is not known to cache", hashValue))
	// }

	Debug.Printf("Cache has key '%s'", hashValue)

	// Key is known, but not loaded into RAM
	// if content == nil {
	Debug.Printf("Cache has content for '%s' already loaded", hashValue)

	content, err := ioutil.ReadFile(c.folder + hashValue)
	if err != nil {
		Error.Printf("Error reading cached file '%s'", hashValue)
		return hashValue, nil, err
	}

	// c.mutex.Lock()
	// c.knownValues[hashValue] = content
	// c.mutex.Unlock()
	// }

	return hashValue, content, nil
}

// Put item into cache
func (c *Cache) put(key string, content []byte) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	hashValue := calcHash(key)
	cachePath := path.Join(c.folder, hashValue)
	err := ioutil.WriteFile(cachePath, content, 0644)

	// Make sure, that the RAM-cache only holds values we were able to write.
	// This is a decision to prevent a false impression of the cache: If the
	// write fails, the cache isn't working correctly, which should be fixed by
	// the user of this cache.
	if err != nil {
		return hashValue, err
	}
	// c.mutex.Lock()
	// c.knownValues[hashValue] = content
	// c.mutex.Unlock()
	Debug.Printf("Cache wrote %s content into '%s'", key, cachePath)

	return hashValue, err
}

func calcHash(data string) string {
	sha := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sha[:])
}
