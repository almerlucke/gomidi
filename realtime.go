package midi

import (
	"fmt"
	"io"
)

// SystemRealTimeEvent real time event
type SystemRealTimeEvent struct {
	coreEvent
}

// String representation
func (e *SystemRealTimeEvent) String() string {
	return fmt.Sprintf("%v: deltaTime %v", eventTypeToString(e.eventType), e.deltaTime)
}

// WriteTo writer
func (e *SystemRealTimeEvent) WriteTo(w io.Writer) (int64, error) {
	var totalBytesWritten int64

	n, err := w.Write(writeVariableLengthValue(e.deltaTime))
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	var statusByte byte

	switch e.eventType {
	case TimingClock:
		statusByte = 0xF8
	case Start:
		statusByte = 0xFA
	case Continue:
		statusByte = 0xFB
	case Stop:
		statusByte = 0xFC
	case ActiveSensing:
		statusByte = 0xFE
	}

	n, err = w.Write([]byte{statusByte})
	if err != nil {
		return 0, err
	}

	return totalBytesWritten + int64(n), nil
}

// DeltaTime of the system real time event
func (e *SystemRealTimeEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the system real time event
func (e *SystemRealTimeEvent) EventType() EventType {
	return e.eventType
}

// parseSystemRealTimeEvent parses a system real time event
func parseSystemRealTimeEvent(deltaTime uint32, eventType EventType) (event Event, bytesRead uint32, err error) {
	event = &SystemRealTimeEvent{
		coreEvent: coreEvent{
			deltaTime: deltaTime,
			eventType: eventType,
		},
	}

	return
}

// parseTimingClock parses a timing clock event
func parseTimingClock(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemRealTimeEvent(deltaTime, TimingClock)
}

// parseStart parses a start event
func parseStart(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemRealTimeEvent(deltaTime, Start)
}

// parseContinue parses a continue event
func parseContinue(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemRealTimeEvent(deltaTime, Continue)
}

// parseStop parses a stop event
func parseStop(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemRealTimeEvent(deltaTime, Stop)
}

// parseActiveSensing parses an active sensing event
func parseActiveSensing(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseSystemRealTimeEvent(deltaTime, ActiveSensing)
}
