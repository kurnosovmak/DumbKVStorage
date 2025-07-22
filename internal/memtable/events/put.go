package events

import (
	"encoding/binary"
	"fmt"
)

type PutEvent struct {
	Key   string
	Value []byte
}

func (e *PutEvent) Type() EventType {
	return EventTypePut
}

func (e *PutEvent) Encode() []byte {
	keyBytes := []byte(e.Key)
	buf := make([]byte, 4+4+len(keyBytes)+4+len(e.Value))

	binary.LittleEndian.PutUint16(buf[0:3], uint16(EventTypePut))
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(keyBytes)))
	copy(buf[8:], keyBytes)
	offset := 8 + len(keyBytes)

	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(e.Value)))
	copy(buf[offset+4:], e.Value)

	return buf
}

func (e *PutEvent) Decode(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("invalid PutEvent")
	}
	eventType := binary.LittleEndian.Uint16(data[0:4])
	if EventType(eventType) != EventTypePut {
		return fmt.Errorf("invalid PutEvent type")
	}
	keyLen := binary.LittleEndian.Uint32(data[4:])
	if len(data) < int(8+keyLen+4) {
		return fmt.Errorf("invalid PutEvent length")
	}
	e.Key = string(data[8 : 8+keyLen])
	valLen := binary.LittleEndian.Uint32(data[8+keyLen:])
	if len(data) < int(8+keyLen+4+valLen) {
		return fmt.Errorf("invalid PutEvent value length")
	}
	e.Value = append([]byte(nil), data[8+keyLen+4:]...)
	return nil
}
