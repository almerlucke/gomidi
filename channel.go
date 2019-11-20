package midi

import (
	"fmt"
	"io"
)

// ChannelEvent represents channel voice and mode messages
type ChannelEvent struct {
	coreEvent
	Channel uint16
	Value1  uint16
	Value2  uint16
}

// String representation
func (e *ChannelEvent) String() string {
	if e.eventType == PitchWheelChange || e.eventType == ProgramChange {
		return fmt.Sprintf("%v: deltaTime %v, channel %v, value %v", eventTypeToString(e.eventType), e.deltaTime, e.Channel, e.Value1)
	}

	return fmt.Sprintf("%v: deltaTime %v, channel %v, value1 %v, value2 %v", eventTypeToString(e.eventType), e.deltaTime, e.Channel, e.Value1, e.Value2)
}

// WriteTo writer
func (e *ChannelEvent) WriteTo(w io.Writer) (int64, error) {
	var totalBytesWritten int64

	n, err := w.Write(writeVariableLengthInteger(e.deltaTime))
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	data := make([]byte, 3)
	data[1] = byte(e.Value1)
	data[2] = byte(e.Value2)

	numBytes := 3

	switch e.eventType {
	case NoteOff:
		data[0] = 0x8
	case NoteOn:
		data[0] = 0x9
	case PolyphonicKeyPressure:
		data[0] = 0xA
	case ControlChange:
		data[0] = 0xB
	case ProgramChange:
		data[0] = 0xC
		numBytes = 2
	case ChannelPressure:
		data[0] = 0xD
		numBytes = 2
	case PitchWheelChange:
		data[0] = 0xE
		data[1] = byte(e.Value1 & 0x7F)
		data[2] = byte(e.Value1 >> 7)
	}

	data[0] = (data[0] << 4) ^ byte(e.Channel)

	n, err = w.Write(data[:numBytes])
	if err != nil {
		return 0, err
	}

	return totalBytesWritten + int64(n), nil
}

// parseChannelEvent parses a channel voice or mode event
func parseChannelEvent(statusByte uint8, deltaTime uint32, eventType EventType, numValues uint8, data []byte) (event Event, bytesRead uint32, err error) {
	ce := &ChannelEvent{}
	ce.deltaTime = deltaTime
	ce.eventType = eventType
	ce.Channel = uint16(statusByte & 0xF)

	if len(data) < int(numValues) {
		err = fmt.Errorf("channel event of type %v expects %v data bytes", eventTypeToString(eventType), numValues)
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

// parseNoteOff parses a note off event
func parseNoteOff(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseChannelEvent(statusByte, deltaTime, NoteOff, 2, data)
}

// parseNoteOn parses a note off event
func parseNoteOn(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseChannelEvent(statusByte, deltaTime, NoteOn, 2, data)
}

// parsePolyphonicKeyPressure parses a polyphonic key pressure event
func parsePolyphonicKeyPressure(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseChannelEvent(statusByte, deltaTime, PolyphonicKeyPressure, 2, data)
}

// parseControlChange parses a control change event
func parseControlChange(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseChannelEvent(statusByte, deltaTime, ControlChange, 2, data)
}

// parseProgramChange parses a program change event
func parseProgramChange(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseChannelEvent(statusByte, deltaTime, ProgramChange, 1, data)
}

// parseChannelPressure parses a channel pressure event
func parseChannelPressure(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return parseChannelEvent(statusByte, deltaTime, ChannelPressure, 1, data)
}

// parsePitchWheelChange parses a pitch wheel change event
func parsePitchWheelChange(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	event, bytesRead, err = parseChannelEvent(statusByte, deltaTime, PitchWheelChange, 2, data)
	if err == nil {
		// Get channel event
		pw := event.(*ChannelEvent)

		// Combine into single 14 bits pitch wheel value
		pw.Value1 = (pw.Value2 << 7) ^ pw.Value1
	}

	return
}
