package memtable

import (
	"fmt"
	"log/slog"
	"sync"
	"testdb/internal/memtable/events"
	"testdb/internal/wal"
)

type MemTable interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, bool)
	Delete(key string) error
	Flush(newWal WriteAheadLog) (map[string][]byte, error)
	Size() (int, error)
	Init() error
}

type WriteAheadLog interface {
	Write(event wal.Encoded) error
	ReadAll(callback func([]byte) error) error
}

type Snapshot interface {
	InitFromSnapshot(callback func(value string, data []byte) error) error
}

type memTable struct {
	mu       *sync.RWMutex
	table    map[string][]byte
	wal      WriteAheadLog
	log      *slog.Logger
	snapshot Snapshot
}

func NewMemTable(log *slog.Logger, wal WriteAheadLog, snapshot Snapshot) MemTable {
	return &memTable{
		mu:       &sync.RWMutex{},
		table:    make(map[string][]byte, 10_000),
		wal:      wal,
		log:      log,
		snapshot: snapshot,
	}
}

func (m *memTable) Init() error {
	const op = "memtable.Init"

	m.mu.Lock()
	defer m.mu.Unlock()

	log := m.log.With(slog.String("op", op))

	err := m.snapshot.InitFromSnapshot(func(value string, data []byte) error {
		m.table[value] = data
		return nil
	})

	if err != nil {
		log.Error("error reading snapshot event", slog.Any("err", err))
		return err
	}

	err = m.wal.ReadAll(func(value []byte) error {
		event, err := events.DecodeEvent(value)
		if err != nil {
			return err
		}
		switch event.Type() {
		case events.EventTypeDelete:
			e := event.(*events.DeleteEvent)
			delete(m.table, e.Key)
			break
		case events.EventTypePut:
			e := event.(*events.PutEvent)
			m.table[e.Key] = e.Value
			break
		default:
			return fmt.Errorf("invalid event type %d", event.Type())
		}
		return nil
	})

	if err != nil {
		log.Error("error reading wal event", slog.Any("err", err))
		return err
	}
	return nil
}

func (m *memTable) Put(key string, value []byte) error {
	const op = "memtable.Put"
	log := m.log.With(slog.String("op", op), slog.String("key", key), slog.String("value", string(value)))

	m.mu.Lock()
	defer m.mu.Unlock()
	err := m.wal.Write(&events.PutEvent{
		Key:   key,
		Value: value,
	})
	if err != nil {
		log.Error("failed to write", slog.Any("err", err))
		return err
	}
	m.table[key] = value
	return nil
}

func (m *memTable) Get(key string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.table[key]
	return val, ok
}

func (m *memTable) Delete(key string) error {
	const op = "memtable.Delete"
	log := m.log.With(slog.String("op", op), slog.String("key", key))

	m.mu.Lock()
	defer m.mu.Unlock()
	err := m.wal.Write(&events.DeleteEvent{
		Key: key,
	})
	if err != nil {
		log.Error("failed to write", slog.Any("err", err))
		return err
	}
	delete(m.table, key)
	return nil
}

func (m *memTable) Flush(newWal WriteAheadLog) (map[string][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// возвращаем копию для безопасного snapshot
	copied := make(map[string][]byte, len(m.table))
	for k, v := range m.table {
		valCopy := make([]byte, len(v))
		copy(valCopy, v)
		copied[k] = valCopy
	}
	m.wal = newWal
	return copied, nil
}

func (m *memTable) Size() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.table), nil
}
