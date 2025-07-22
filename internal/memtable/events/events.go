package events

import (
	"encoding/binary"
	"fmt"
)

type EventType uint16

const (
	EventTypePut    EventType = 1
	EventTypeDelete EventType = 2
)

type Event interface {
	Encode() []byte
	Type() EventType
}

type EventDecoder interface {
	Decode([]byte) error
}

func DecodeEvent(data []byte) (Event, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty event")
	}
	var ev EventDecoder

	eventType := binary.LittleEndian.Uint16(data[0:3])

	switch EventType(eventType) {
	case EventTypePut:
		ev = &PutEvent{}
	case EventTypeDelete:
		ev = &DeleteEvent{}
	default:
		return nil, fmt.Errorf("unknown event type: %d", data[0])
	}

	if err := ev.Decode(data); err != nil {
		return nil, err
	}
	return ev.(Event), nil
}
