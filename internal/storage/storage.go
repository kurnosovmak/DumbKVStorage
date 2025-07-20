package storage

import (
	"fmt"
	"sync"
	"vdb/internal/models"
	"vdb/internal/file"
)

type Storage struct {
	mu sync.RWMutex
	db map[string]string
}

func New() *Storage {
	return &Storage{
		db: make(map[string]string),
	}
}

func (s *Storage) Set(kv models.KeyValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db[kv.Key.Key] = kv.Value.Value
	err := file.DumpMapToFile(s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Get(key models.Key) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.db[key.Key]
	return val, ok
}

func (s *Storage) Init() (error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := file.LoadMapFromFile()
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *Storage) Dump() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("%s", s.db)
}