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

// String representation
func (e *SystemCommonEvent) String() string {
	return fmt.Sprintf("%v: deltaTime %v", eventTypeToString(e.eventType), e.deltaTime)
}

// WriteTo writer
func (e *SystemCommonEvent) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

// DeltaTime of the system common event
func (e *SystemCommonEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the system common event
func (e *SystemCommonEvent) EventType() EventType {
	return e.eventType
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
