package midi

import (
	"errors"
	"fmt"
	"io"
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

// String representation
func (e *MetaEvent) String() string {
	return fmt.Sprintf("%v: deltaTime %v, type %v, content %v", eventTypeToString(e.eventType), e.deltaTime, metaTypeToString(e.MetaType), string(e.Data))
}

// WriteTo writer
func (e *MetaEvent) WriteTo(w io.Writer) (int64, error) {
	var totalBytesWritten int64

	n, err := w.Write(writeVariableLengthInteger(e.deltaTime))
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	n, err = w.Write([]byte{0xFF})
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	var metaType byte

	switch e.MetaType {
	case SequenceNumber:
		metaType = 0x0
	case Text:
		metaType = 0x1
	case CopyrightNotice:
		metaType = 0x2
	case TrackName:
		metaType = 0x3
	case InstrumentName:
		metaType = 0x4
	case Lyric:
		metaType = 0x5
	case Marker:
		metaType = 0x6
	case CuePoint:
		metaType = 0x7
	case ChannelPrefix:
		metaType = 0x20
	case EndOfTrack:
		metaType = 0x2F
	case SetTempo:
		metaType = 0x51
	case SMPTEOffset:
		metaType = 0x54
	case TimeSignature:
		metaType = 0x58
	case KeySignature:
		metaType = 0x59
	case SequencerSpecific:
		metaType = 0x7F
	}

	n, err = w.Write([]byte{metaType})
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	lengthData := writeVariableLengthInteger(uint32(len(e.Data)))
	n, err = w.Write(lengthData)
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	n, err = w.Write(e.Data)
	if err != nil {
		return 0, err
	}

	return totalBytesWritten + int64(n), nil
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
