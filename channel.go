package midi

import (
	"fmt"
)

// ChannelEvent represents channel voice and mode messages
type ChannelEvent struct {
	coreEvent
	Channel uint16
	Value1  uint16
	Value2  uint16
}

// DeltaTime of the channel event
func (e *ChannelEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the channel event
func (e *ChannelEvent) EventType() EventType {
	return e.eventType
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
