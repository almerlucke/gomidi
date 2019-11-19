package midi

import (
	"errors"
)

// MetaType to identify meta events
type MetaType uint8

const (
	// SequenceNumber meta event
	SequenceNumber MetaType = 0x0
	// Text meta event
	Text MetaType = 0x1
	// CopyrightNotice meta event
	CopyrightNotice MetaType = 0x2
	// TrackName meta event
	TrackName MetaType = 0x3
	// InstrumentName meta event
	InstrumentName MetaType = 0x4
	// Lyric meta event
	Lyric MetaType = 0x5
	// Marker meta event
	Marker MetaType = 0x6
	// CuePoint meta event
	CuePoint MetaType = 0x7
	// ChannelPrefix meta event
	ChannelPrefix MetaType = 0x20
	// EndOfTrack meta event
	EndOfTrack MetaType = 0x2F
	// SetTempo meta event
	SetTempo MetaType = 0x51
	// SMPTEOffset meta event
	SMPTEOffset MetaType = 0x54
	// TimeSignature meta event
	TimeSignature MetaType = 0x58
	// KeySignature meta event
	KeySignature MetaType = 0x59
	// SequencerSpecific meta event
	SequencerSpecific MetaType = 0x7F
)

// MetaEvent struct for all meta events
type MetaEvent struct {
	coreEvent
	MetaType MetaType
	Data     []byte
}

// DeltaTime of the meta event
func (e *MetaEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the meta event
func (e *MetaEvent) EventType() EventType {
	return e.eventType
}

// metaTypeToString converts a type to a string for debugging
func metaTypeToString(metaType MetaType) string {
	switch metaType {
	case SequenceNumber:
		return "SequenceNumber"
	case Text:
		return "Text"
	case CopyrightNotice:
		return "CopyrightNotice"
	case TrackName:
		return "TrackName"
	case InstrumentName:
		return "InstrumentName"
	case Lyric:
		return "Lyric"
	case Marker:
		return "Marker"
	case CuePoint:
		return "CuePoint"
	case ChannelPrefix:
		return "ChannelPrefix"
	case EndOfTrack:
		return "EndOfTrack"
	case SetTempo:
		return "SetTempo"
	case SMPTEOffset:
		return "SMPTEOffset"
	case TimeSignature:
		return "TimeSignature"
	case KeySignature:
		return "KeySignature"
	case SequencerSpecific:
		return "SequencerSpecific"
	}

	return "Unknown"
}

// parseMeta parses a meta event
func parseMeta(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	if len(data) == 0 {
		err = errors.New("end of data before meta event was identified")
		return
	}

	// Get meta data type from byte
	metaType := MetaType(data[0])

	// Skip to next byte
	data = data[1:]

	// Get variable length num bytes
	numBytes, bytesRead, err := readVariableLengthInteger(data)
	if err != nil {
		return
	}

	// Skip bytes for variable length value
	data = data[bytesRead:]
	if uint32(len(data)) < numBytes {
		err = errors.New("given meta event length exceeds available data length")
		return
	}

	bytesRead += numBytes

	// Copy meta data
	metaData := make([]byte, numBytes)
	copy(metaData, data)

	// Create new event
	event = &MetaEvent{
		coreEvent: coreEvent{
			eventType: Meta,
			deltaTime: deltaTime,
		},
		MetaType: metaType,
		Data:     metaData,
	}

	// Offset 1 for metaStatusByte
	bytesRead++

	return
}
