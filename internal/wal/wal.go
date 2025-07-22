package wal

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
)

const (
	WalFilePrefix = "test_"
	WalFileExt    = ".wal"
)

type Wal struct {
	mu   sync.Mutex
	file *os.File
	path string
}

type Encoded interface {
	Encode() []byte
}

func NewWal(path string) (*Wal, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &Wal{
		file: file,
		path: path,
	}, nil
}

// Write записывает одну запись в формате: int32(len) + data
func (w *Wal) Write(encoded Encoded) error {
	data := encoded.Encode()
	w.mu.Lock()
	defer w.mu.Unlock()

	length := int32(len(data))
	if err := binary.Write(w.file, binary.LittleEndian, length); err != nil {
		return err
	}
	if _, err := w.file.Write(data); err != nil {
		return err
	}
	return w.file.Sync()
}

// ReadAll читает все записи WAL
func (w *Wal) ReadAll(callback func([]byte) error) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for {
		var l int32
		err := binary.Read(w.file, binary.LittleEndian, &l)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buf := make([]byte, l)
		if _, err := io.ReadFull(w.file, buf); err != nil {
			return err
		}
		err = callback(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close закрывает WAL-файл
func (w *Wal) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
