package main

import (
	"encoding/json"
	"os"
	"sync"
)

// DB is an interface that represents a key-value database
// NOTE: key-value storage chosen because it provides fast O(1) read complexity.
// This implementation have tradeoff with memory, so if large amount of records needed, other storage should be considered as an option.
type DB interface {
	Set(key string, value string) error
	Get(key string) (string, bool)
}

func NewFileDB(filename string) (DB, error) {
	db := &fileDB{
		filename: filename,
		data:     make(map[string]string),
	}
	err := db.loadFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return db, nil
}

// fileDB is a file-based key-value storage
// NOTE: fileDB use file as back storage for writes, read operation use InMemory data,
// this can be easily changes, but I decided to avoid file operations on read.
type fileDB struct {
	filename string
	mutex    sync.RWMutex
	data     map[string]string
}

func (s *fileDB) Set(key string, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
	return s.writeFile()
}

func (s *fileDB) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.data[key]
	return value, ok
}

// loadFile loads writeJSON content to in-memory storage
func (s *fileDB) loadFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := os.ReadFile(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, initialize with an empty map
			s.data = make(map[string]string)
			return nil
		}
		return err
	}

	return json.Unmarshal(file, &s.data)
}

// writeFile flushes in-memory storage to file to persist
func (s *fileDB) writeFile() error {
	file, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filename, file, 0644)
}
