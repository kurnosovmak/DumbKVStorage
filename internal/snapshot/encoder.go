package snapshot

import (
	"encoding/binary"
	"io"
)

func write(writer io.Writer, data map[string][]byte) error {
	for key, value := range data {
		if err := writeBytes(writer, []byte(key)); err != nil {
			return err
		}
		if err := writeBytes(writer, value); err != nil {
			return err
		}
	}
	return nil
}

func read(reader io.Reader, callback func(key string, value []byte) error) error {
	key, err := readBytes(reader)
	if err != nil {
		return err
	}
	value, err := readBytes(reader)
	if err != nil {
		return err
	}
	//fmt.Println(string(key), string(value))
	err = callback(string(key), value)
	if err != nil {
		return err
	}
	return nil
}

func writeBytes(w io.Writer, s []byte) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}
	_, err := w.Write(s)
	return err
}

func readBytes(r io.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return []byte(""), err
	}
	buf := make([]byte, length)
	_, err := io.ReadFull(r, buf)
	return buf, err
}
