package midi

import (
	"fmt"
	"io"
)

// SystemCommonEvent represents a system common message
type SystemCommonEvent struct {
	coreEvent
	Value1 uint16
	Value2 uint16
}

// WriteTo writer
func (e *SystemCommonEvent) WriteTo(w io.Writer) (int64, error) {
	var totalBytesWritten int64

	n, err := w.Write(writeVariableLengthInteger(e.deltaTime))
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	data := make([]byte, 3)
	numBytes := 1

	data[1] = byte(e.Value1)
	data[2] = byte(e.Value2)

	switch e.eventType {
	case SongPositionPointer:
		data[0] = 0xF2
		data[1] = byte(e.Value1 & 0x7F)
		data[2] = byte(e.Value1 >> 7)
		numBytes = 3
	case SongSelect:
		data[0] = 0xF3
		numBytes = 2
	case TuneRequest:
		data[0] = 0xF6
	}

	data = data[:numBytes]

	n, err = w.Write(data)
	if err != nil {
		return 0, err
	}

	return totalBytesWritten + int64(n), nil
}

// parseSystemCommonEvent parses a system common event
func parseSystemCommonEvent(deltaTime uint32, eventType EventType, numValues uint8, data []byte) (event Event, bytesRead uint32, err error) {
	ce := &SystemCommonEvent{}
	ce.deltaTime = deltaTime
	ce.eventType = eventType

	if len(data) < int(numValues) {
		err = fmt.Errorf("system common event of type %v expects %v data bytes", eventTypeToString(eventType), numValues)
		return
	}

	if numValues == 1 {
		ce.Value1 = uint16(data[0])
	} else if numValues == 2 {
		ce.Value1 = uint16(data[0])
		ce.Value2 = uint16(data[1])
	}

	bytesRead = uint32(numValues)
	event = ce

	return
}

// parseSongPositionPointer parses a song position pointer event
func parseSongPositionPointer(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	event, bytesRead, err = parseSystemCommonEvent(deltaTime, SongPositionPointer, 2, data)
	if err == nil {
		// Get system common event
		pw := event.(*SystemCommonEvent)

		// Combine into single 14 bits song position pointer
		pw.Value1 = (pw.Value2 << 7) ^ pw.Value1
	}

	return
}

// parseSongSelect parses a song select event
func parseSongSelect(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemCommonEvent(deltaTime, SongSelect, 1, data)
}

// parseTuneRequest parses a tune request
func parseTuneRequest(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemCommonEvent(deltaTime, TuneRequest, 0, data)
}
