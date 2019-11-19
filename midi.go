package midi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

// ChunkType indicates the type of chunk we are dealing with
type ChunkType string

// Format type
type Format uint16

// DivisionType midi delta time
type DivisionType uint8

// EventType to identify midi events
type EventType uint8

// MetaType to identify meta events
type MetaType uint8

// Chunk holds midi chunk information
type Chunk struct {
	io.WriterTo
	io.ReaderFrom
	Type   ChunkType
	Length uint32
	Data   []byte
}

// HeaderInfo holds midi file header info
type HeaderInfo struct {
	Format              Format
	NumTracks           uint16
	Division            uint16
	DivisionType        DivisionType
	TicksPerQuarterNote uint16
	FramesPerSecond     uint8
	TicksPerFrame       uint8
}

// File holds all midi chunks and other info
type File struct {
	io.WriterTo
	io.ReaderFrom
	Info   *HeaderInfo
	Chunks []*Chunk
}

// Event interface for all midi events
type Event interface {
	DeltaTime() uint32
	EventType() EventType
}

// CoreEvent to include by other event structs to satisfy Event interface
type CoreEvent struct {
	deltaTime uint32
	eventType EventType
}

// MetaEvent struct for all meta events
type MetaEvent struct {
	CoreEvent
	MetaType MetaType
	Data     []byte
}

// SystemRealTimeEvent real time event
type SystemRealTimeEvent struct {
	CoreEvent
}

// SystemExclusiveEvent representation
type SystemExclusiveEvent struct {
	CoreEvent
	Data []byte
}

// ChannelEvent represents channel voice and mode messages
type ChannelEvent struct {
	CoreEvent
	Channel uint16
	Value1  uint16
	Value2  uint16
}

// SystemCommonEvent represents a system common message
type SystemCommonEvent struct {
	CoreEvent
	Value1 uint16
	Value2 uint16
}

// ParseFunction type
type ParseFunction func(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error)

const (
	// HeaderType indicates a midi header chunk
	HeaderType ChunkType = "MThd"
	// TrackType indicates a midi track chunk
	TrackType ChunkType = "MTrk"
)

const (
	// Format0 midi file
	Format0 Format = 0
	// Format1 midi file
	Format1 Format = 1
	// Format2 midi file
	Format2 Format = 2
)

const (
	// DivisionTicksPerQuarterNote division
	DivisionTicksPerQuarterNote DivisionType = 0
	// DivisionFramesTicks division
	DivisionFramesTicks DivisionType = 1
)

const (
	// NoteOff midi event
	NoteOff EventType = iota
	// NoteOn midi event
	NoteOn
	// PolyphonicKeyPressure midi event
	PolyphonicKeyPressure
	// ControlChange midi event
	ControlChange
	// ProgramChange midi event
	ProgramChange
	// ChannelPressure midi event
	ChannelPressure
	// PitchWheelChange midi event
	PitchWheelChange
	// SystemExclusive midi event
	SystemExclusive
	// SongPositionPointer midi event
	SongPositionPointer
	// SongSelect midi event
	SongSelect
	// TuneRequest midi event
	TuneRequest
	// TimingClock midi event
	TimingClock
	// Start midi event
	Start
	// Continue midi event
	Continue
	// Stop midi event
	Stop
	// ActiveSensing midi event
	ActiveSensing
	// Meta midi event
	Meta
)

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

var eventTypeToParseFunctionMapping = map[EventType]ParseFunction{
	NoteOff:               ParseNoteOff,
	NoteOn:                ParseNoteOn,
	PolyphonicKeyPressure: ParsePolyphonicKeyPressure,
	ControlChange:         ParseControlChange,
	ProgramChange:         ParseProgramChange,
	ChannelPressure:       ParseChannelPressure,
	PitchWheelChange:      ParsePitchWheelChange,
	SystemExclusive:       ParseSystemExclusive,
	SongPositionPointer:   ParseSongPositionPointer,
	SongSelect:            ParseSongSelect,
	TuneRequest:           ParseTuneRequest,
	TimingClock:           ParseTimingClock,
	Start:                 ParseStart,
	Continue:              ParseContinue,
	Stop:                  ParseStop,
	ActiveSensing:         ParseActiveSensing,
	Meta:                  ParseMeta,
}

// DeltaTime of the meta event
func (e *MetaEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the meta event
func (e *MetaEvent) EventType() EventType {
	return e.eventType
}

// DeltaTime of the system real time event
func (e *SystemRealTimeEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the system real time event
func (e *SystemRealTimeEvent) EventType() EventType {
	return e.eventType
}

// DeltaTime of the system exclusive event
func (e *SystemExclusiveEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the system exclusive event
func (e *SystemExclusiveEvent) EventType() EventType {
	return e.eventType
}

// DeltaTime of the channel event
func (e *ChannelEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the channel event
func (e *ChannelEvent) EventType() EventType {
	return e.eventType
}

// DeltaTime of the system common event
func (e *SystemCommonEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// EventType of the system common event
func (e *SystemCommonEvent) EventType() EventType {
	return e.eventType
}

func eventTypeToString(eventType EventType) string {
	switch eventType {
	case NoteOff:
		return "NoteOff"
	case NoteOn:
		return "NoteOn"
	case PolyphonicKeyPressure:
		return "PolyphonicKeyPressure"
	case ControlChange:
		return "ControlChange"
	case ProgramChange:
		return "ProgramChange"
	case ChannelPressure:
		return "ChannelPressure"
	case PitchWheelChange:
		return "PitchWheelChange"
	case SystemExclusive:
		return "SystemExclusive"
	case SongPositionPointer:
		return "SongPositionPointer"
	case SongSelect:
		return "SongSelect"
	case TuneRequest:
		return "TuneRequest"
	case TimingClock:
		return "TimingClock"
	case Start:
		return "Start"
	case Continue:
		return "Continue"
	case Stop:
		return "Stop"
	case ActiveSensing:
		return "ActiveSensing"
	case Meta:
		return "Meta"
	}

	return ""
}

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

	return ""
}

// ReadVariableLengthInteger reads a variable length integer from a slice of bytes
func ReadVariableLengthInteger(bs []byte) (result uint32, bytesRead uint32, err error) {
	foundZero := false
	err = nil

	for _, b := range bs {
		bytesRead++
		result <<= 7
		result ^= uint32(b) & 0x7F
		if (b >> 7) == 0 {
			foundZero = true
			break
		}
	}

	if !foundZero {
		return 0, 0, errors.New("a variable length quantity should end with a byte with the most significant bit set to 0")
	}

	return
}

// ParseChannelEvent parses a channel voice or mode event
func ParseChannelEvent(statusByte uint8, deltaTime uint32, eventType EventType, numValues uint8, data []byte) (event Event, bytesRead uint32, err error) {
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

// ParseNoteOff parses a note off event
func ParseNoteOff(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseChannelEvent(statusByte, deltaTime, NoteOff, 2, data)
}

// ParseNoteOn parses a note off event
func ParseNoteOn(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseChannelEvent(statusByte, deltaTime, NoteOn, 2, data)
}

// ParsePolyphonicKeyPressure parses a polyphonic key pressure event
func ParsePolyphonicKeyPressure(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseChannelEvent(statusByte, deltaTime, PolyphonicKeyPressure, 2, data)
}

// ParseControlChange parses a control change event
func ParseControlChange(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseChannelEvent(statusByte, deltaTime, ControlChange, 2, data)
}

// ParseProgramChange parses a program change event
func ParseProgramChange(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseChannelEvent(statusByte, deltaTime, ProgramChange, 1, data)
}

// ParseChannelPressure parses a channel pressure event
func ParseChannelPressure(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseChannelEvent(statusByte, deltaTime, ChannelPressure, 1, data)
}

// ParsePitchWheelChange parses a pitch wheel change event
func ParsePitchWheelChange(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	event, bytesRead, err = ParseChannelEvent(statusByte, deltaTime, PitchWheelChange, 2, data)
	if err == nil {
		// Get channel event
		pw := event.(*ChannelEvent)

		// Combine into single 14 bits pitch wheel value
		pw.Value1 = (pw.Value2 << 7) ^ pw.Value1
	}

	return
}

// ParseSystemExclusive parses a system exclusive event
func ParseSystemExclusive(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	numBytes, bytesRead, err := ReadVariableLengthInteger(data)
	if err != nil {
		return
	}

	data = data[bytesRead:]
	if uint32(len(data)) < numBytes {
		err = errors.New("given system exclusive event length exceeds available data length")
		return
	}

	bytesRead += numBytes
	exclusiveData := make([]byte, numBytes)

	copy(exclusiveData, data)

	event = &SystemExclusiveEvent{
		CoreEvent: CoreEvent{
			deltaTime: deltaTime,
			eventType: SystemExclusive,
		},
		Data: exclusiveData,
	}

	return
}

// ParseSystemCommonEvent parses a system common event
func ParseSystemCommonEvent(deltaTime uint32, eventType EventType, numValues uint8, data []byte) (event Event, bytesRead uint32, err error) {
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

// ParseSongPositionPointer parses a song position pointer event
func ParseSongPositionPointer(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	event, bytesRead, err = ParseSystemCommonEvent(deltaTime, SongPositionPointer, 2, data)
	if err == nil {
		// Get system common event
		pw := event.(*SystemCommonEvent)

		// Combine into single 14 bits song position pointer
		pw.Value1 = (pw.Value2 << 7) ^ pw.Value1
	}

	return
}

// ParseSongSelect parses a song select event
func ParseSongSelect(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemCommonEvent(deltaTime, SongSelect, 1, data)
}

// ParseTuneRequest parses a tune request
func ParseTuneRequest(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemCommonEvent(deltaTime, TuneRequest, 0, data)
}

// ParseSystemRealTimeEvent parses a system real time event
func ParseSystemRealTimeEvent(deltaTime uint32, eventType EventType) (event Event, bytesRead uint32, err error) {
	event = &SystemRealTimeEvent{
		CoreEvent: CoreEvent{
			deltaTime: deltaTime,
			eventType: eventType,
		},
	}

	return
}

// ParseTimingClock parses a timing clock event
func ParseTimingClock(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemRealTimeEvent(deltaTime, TimingClock)
}

// ParseStart parses a start event
func ParseStart(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemRealTimeEvent(deltaTime, Start)
}

// ParseContinue parses a continue event
func ParseContinue(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemRealTimeEvent(deltaTime, Continue)
}

// ParseStop parses a stop event
func ParseStop(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemRealTimeEvent(deltaTime, Stop)
}

// ParseActiveSensing parses an active sensing event
func ParseActiveSensing(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	return ParseSystemRealTimeEvent(deltaTime, ActiveSensing)
}

// ParseMeta parse a meta event
func ParseMeta(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	if len(data) == 0 {
		err = errors.New("end of data before meta event was identified")
		return
	}

	metaType := MetaType(data[0])
	data = data[1:]

	log.Printf("meta %v\n", metaTypeToString(metaType))

	numBytes, bytesRead, err := ReadVariableLengthInteger(data)
	if err != nil {
		return
	}

	data = data[bytesRead:]
	if uint32(len(data)) < numBytes {
		err = errors.New("given meta event length exceeds available data length")
		return
	}

	bytesRead += numBytes

	metaData := make([]byte, numBytes)

	copy(metaData, data)

	event = &MetaEvent{
		CoreEvent: CoreEvent{
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

// Events get events from chunk
func (c *Chunk) Events() ([]Event, error) {
	runningStatusActive := false
	var runningStatusByte uint8
	data := c.Data
	events := []Event{}

	for {
		deltaTime, bytesRead, err := ReadVariableLengthInteger(data)
		if err != nil {
			return nil, err
		}

		data = data[bytesRead:]

		if len(data) == 0 {
			return nil, errors.New("expected another event after delta time")
		}

		statusByte := data[0]

		if (statusByte >> 7) == 1 {
			// Skip status byte
			data = data[1:]
		} else {
			// Data byte, we expect runningStatusActive to be true
			if !runningStatusActive {
				return nil, errors.New("received data byte without running status active")
			}

			statusByte = runningStatusByte
		}

		var parseFunc ParseFunction
		var event Event

		switch {
		case (statusByte >> 4) == 0x8:
			parseFunc = eventTypeToParseFunctionMapping[NoteOff]
			runningStatusActive = true
			runningStatusByte = statusByte
		case (statusByte >> 4) == 0x9:
			parseFunc = eventTypeToParseFunctionMapping[NoteOn]
			runningStatusActive = true
			runningStatusByte = statusByte
		case (statusByte >> 4) == 0xA:
			parseFunc = eventTypeToParseFunctionMapping[PolyphonicKeyPressure]
			runningStatusActive = true
			runningStatusByte = statusByte
		case (statusByte >> 4) == 0xB:
			parseFunc = eventTypeToParseFunctionMapping[ControlChange]
			runningStatusActive = true
			runningStatusByte = statusByte
		case (statusByte >> 4) == 0xC:
			parseFunc = eventTypeToParseFunctionMapping[ProgramChange]
			runningStatusActive = true
			runningStatusByte = statusByte
		case (statusByte >> 4) == 0xD:
			parseFunc = eventTypeToParseFunctionMapping[ChannelPressure]
			runningStatusActive = true
			runningStatusByte = statusByte
		case (statusByte >> 4) == 0xE:
			parseFunc = eventTypeToParseFunctionMapping[PitchWheelChange]
			runningStatusActive = true
			runningStatusByte = statusByte
		case statusByte == 0xF0:
			parseFunc = eventTypeToParseFunctionMapping[SystemExclusive]
			runningStatusActive = false
		case statusByte == 0xF2:
			parseFunc = eventTypeToParseFunctionMapping[SongPositionPointer]
			runningStatusActive = false
		case statusByte == 0xF3:
			parseFunc = eventTypeToParseFunctionMapping[SongSelect]
			runningStatusActive = false
		case statusByte == 0xF6:
			parseFunc = eventTypeToParseFunctionMapping[TuneRequest]
			runningStatusActive = false
		case statusByte == 0xF7:
			parseFunc = eventTypeToParseFunctionMapping[SystemExclusive]
			runningStatusActive = false
		case statusByte == 0xF8:
			parseFunc = eventTypeToParseFunctionMapping[TimingClock]
		case statusByte == 0xFA:
			parseFunc = eventTypeToParseFunctionMapping[Start]
		case statusByte == 0xFB:
			parseFunc = eventTypeToParseFunctionMapping[Continue]
		case statusByte == 0xFC:
			parseFunc = eventTypeToParseFunctionMapping[Stop]
		case statusByte == 0xFE:
			parseFunc = eventTypeToParseFunctionMapping[ActiveSensing]
		case statusByte == 0xFF:
			parseFunc = eventTypeToParseFunctionMapping[Meta]
		default:
			return nil, fmt.Errorf("unknown status byte %X encountered", statusByte)
		}

		event, bytesRead, err = parseFunc(statusByte, deltaTime, data)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
		data = data[bytesRead:]

		if len(data) == 0 {
			break
		}
	}

	return events, nil
}

// WriteTo writes a chunk to writer
func (c *Chunk) WriteTo(w io.Writer) (int64, error) {
	// Length needs to be written as big endian
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, c.Length)

	n1, err := w.Write([]byte(c.Type))
	if err != nil {
		return 0, err
	}

	n2, err := w.Write(b)
	if err != nil {
		return 0, err
	}

	n3, err := w.Write(c.Data)
	if err != nil {
		return 0, err
	}

	return int64(n1) + int64(n2) + int64(n3), nil
}

// ReadFrom reads chunk data from reader
func (c *Chunk) ReadFrom(r io.Reader) (int64, error) {
	var n int64

	p := make([]byte, 4)
	nb, err := r.Read(p)
	if err != nil {
		return 0, err
	}

	n += int64(nb)

	c.Type = ChunkType(p)
	err = binary.Read(r, binary.BigEndian, &c.Length)
	if err != nil {
		return 0, err
	}

	c.Data = make([]byte, c.Length)
	nb, err = r.Read(c.Data)
	if err != nil {
		return 0, err
	}

	n += int64(nb)

	return n, nil
}

// HeaderInfo returns header info
func (c *Chunk) HeaderInfo() *HeaderInfo {
	info := &HeaderInfo{}

	info.Format = Format(binary.BigEndian.Uint16(c.Data))
	info.NumTracks = binary.BigEndian.Uint16(c.Data[2:])
	info.Division = binary.BigEndian.Uint16(c.Data[4:])

	if (info.Division >> 15) == 1 {
		info.DivisionType = DivisionFramesTicks
		info.FramesPerSecond = uint8((info.Division & 0x7FFF) >> 8)
		info.TicksPerFrame = uint8(info.Division & 0xFF)
	} else {
		info.DivisionType = DivisionTicksPerQuarterNote
		info.TicksPerQuarterNote = info.Division
	}

	return info
}

// ReadFrom reads a midi file from reader
func (mf *File) ReadFrom(r io.Reader) (int64, error) {
	var n int64

	mf.Chunks = []*Chunk{}

	for {
		chunk := &Chunk{}
		nb, err := chunk.ReadFrom(r)
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, err
		}

		if chunk.Type == HeaderType {
			if chunk.Length != 6 {
				return 0, errors.New("midi header chunk data should be 6 bytes long")
			}

			mf.Info = chunk.HeaderInfo()
			mf.Chunks = append(mf.Chunks, chunk)
		} else if chunk.Type == TrackType {
			mf.Chunks = append(mf.Chunks, chunk)
		}

		n += int64(nb)
	}

	if mf.Info == nil {
		return 0, errors.New("no midi header chunk found")
	}

	return n, nil
}

// WriteTo writes a file to writer
func (mf *File) WriteTo(w io.Writer) (int64, error) {
	var n int64

	for _, chunk := range mf.Chunks {
		nb, err := chunk.WriteTo(w)
		if err != nil {
			return 0, nil
		}

		n += nb
	}

	return n, nil
}
