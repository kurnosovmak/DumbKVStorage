package snapshot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	HeaderFileName     = "header.bin"
	SnapshotFilePrefix = "snapshot_"
	SnapshotFileExt    = ".bin"
	MinSnapshotNumber  = 1
)

type Header struct {
	mu                    *sync.Mutex
	folder                string
	currentWalFileNumber  uint64
	currentSnapshotNumber uint64
}

type Snapshot struct {
	walCounterEvents uint64
	data             map[string][]byte
}

func Init(folder string) (Header, error) {
	if err := os.MkdirAll(folder, 0755); err != nil {
		return Header{}, err
	}
	file, err := os.OpenFile(folder+HeaderFileName, os.O_RDONLY, 0664)
	if err != nil && !os.IsNotExist(err) {
		return Header{}, err
	}
	defer file.Close()
	header := Header{}
	if !os.IsNotExist(err) {
		header, err = readHeader(file)
		if err != nil {
			return Header{}, err
		}
	}
	header.folder = folder
	header.mu = &sync.Mutex{}
	return header, nil

}

func (header *Header) GetWalFileNumber() uint64 {
	header.mu.Lock()
	defer header.mu.Unlock()
	return header.currentWalFileNumber
}

func (header *Header) GetNextWalFileNumber() uint64 {
	header.mu.Lock()
	defer header.mu.Unlock()
	return header.currentWalFileNumber + 1
}

func (header *Header) InitFromSnapshot(callback func(value string, data []byte) error) error {
	if header.currentSnapshotNumber < MinSnapshotNumber {
		return nil
	}
	fileName := fmt.Sprintf("%s%s%d%s", header.folder, SnapshotFilePrefix, header.currentSnapshotNumber, SnapshotFileExt)
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer file.Close()
	bufReader := bufio.NewReaderSize(file, 64*1024) // 1024 KB буфер -> 1 mb
	err = nil
	for err == nil {
		err = read(bufReader, callback)
	}

	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (header *Header) CreateSnapshot(newWalNumber uint64, data map[string][]byte) error {
	header.mu.Lock()
	defer header.mu.Unlock()
	newHeader := Header{
		mu:                    header.mu,
		currentWalFileNumber:  header.currentWalFileNumber,
		currentSnapshotNumber: header.currentSnapshotNumber,
		folder:                header.folder,
	}
	newHeader.currentWalFileNumber = newWalNumber
	newHeader.currentSnapshotNumber++
	if newHeader.currentSnapshotNumber < MinSnapshotNumber {
		newHeader.currentSnapshotNumber = MinSnapshotNumber
	}

	fileName := fmt.Sprintf("%s%s%d%s", newHeader.folder, SnapshotFilePrefix, newHeader.currentSnapshotNumber, SnapshotFileExt)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	err = write(file, data)
	if err != nil {
		os.Remove(fileName)
		return err
	}

	header.currentWalFileNumber = newHeader.currentWalFileNumber
	header.currentSnapshotNumber = newHeader.currentSnapshotNumber

	file, err = os.OpenFile(header.folder+HeaderFileName, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	err = saveHeader(*header, file)
	if err != nil {
		return err
	}

	return nil
}

func readHeader(reader io.Reader) (Header, error) {
	var currentWalFileNumber, currentSnapshotNumber uint64
	err := binary.Read(reader, binary.LittleEndian, &currentWalFileNumber)
	if err != nil {
		return Header{}, err
	}
	err = binary.Read(reader, binary.LittleEndian, &currentSnapshotNumber)
	if err != nil {
		return Header{}, err
	}

	return Header{
		currentWalFileNumber:  currentWalFileNumber,
		currentSnapshotNumber: currentSnapshotNumber,
	}, nil
}

func saveHeader(header Header, writer io.Writer) error {
	err := binary.Write(writer, binary.LittleEndian, header.currentWalFileNumber)
	if err != nil {
		return err
	}
	err = binary.Write(writer, binary.LittleEndian, header.currentSnapshotNumber)
	if err != nil {
		return err
	}
	return nil
}
