package events

import (
	"encoding/binary"
	"fmt"
)

type DeleteEvent struct {
	Key string
}

func (e *DeleteEvent) Type() EventType {
	return EventTypeDelete
}

func (e *DeleteEvent) Encode() []byte {
	keyBytes := []byte(e.Key)
	buf := make([]byte, 4+4+len(keyBytes))

	binary.LittleEndian.PutUint16(buf[0:3], uint16(EventTypeDelete))
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(keyBytes)))
	copy(buf[8:], keyBytes)

	return buf
}

func (e *DeleteEvent) Decode(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("invalid DeleteEvent")
	}
	eventType := binary.LittleEndian.Uint16(data[0:3])
	if EventType(eventType) != EventTypeDelete {
		return fmt.Errorf("invalid DeleteEvent type")
	}
	keyLen := binary.LittleEndian.Uint32(data[4:])
	if len(data) < int(8+keyLen) {
		return fmt.Errorf("invalid DeleteEvent length")
	}
	e.Key = string(data[8 : 8+keyLen])
	return nil
}
