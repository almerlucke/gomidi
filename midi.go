package midi

import (
	"fmt"
	"io"
)

// ChunkType of the chunk
type ChunkType string

// Format of the midi file
type Format uint16

// DivisionType of the header chunk
type DivisionType uint8

// EventType to identify midi events
type EventType uint8

// Chunk separates the midi file in parts
type Chunk struct {
	io.WriterTo
	io.ReaderFrom
	Type   ChunkType
	Length uint32
	Data   []byte
}

// FileHeader is mandatory in a midi file and holds information on number of tracks and tempo
type FileHeader struct {
	Format              Format
	NumTracks           uint16
	Division            uint16
	DivisionType        DivisionType
	TicksPerQuarterNote uint16
	FramesPerSecond     uint8
	TicksPerFrame       uint8
}

// Track contains the midi events (messages)
type Track struct {
	Events []Event
}

// File contains the header, tracks and raw midi chunks, can be used for reading and writing
type File struct {
	io.WriterTo
	io.ReaderFrom
	// Global file info
	Header *FileHeader
	// All tracks in the order they appeared
	Tracks []*Track
	// Also keep a pointer to the raw chunks
	Chunks []*Chunk
}

// NewFile creates a new initialized file
func NewFile() *File {
	return &File{
		Chunks: []*Chunk{},
		Tracks: []*Track{},
	}
}

// Event is the minimal interface all midi event types should conform to
type Event interface {
	io.WriterTo
	fmt.Stringer
	DeltaTime() uint32
	SetDeltaTime(uint32)
	EventType() EventType
}

// coreEvent to include by other event structs to be able to satisfy Event interface
type coreEvent struct {
	deltaTime uint32
	eventType EventType
}

// String generates default event string
func (e *coreEvent) String() string {
	return fmt.Sprintf("%v: deltaTime %v", eventTypeToString(e.eventType), e.deltaTime)
}

// DeltaTime returns 'private' deltatime
func (e *coreEvent) DeltaTime() uint32 {
	return e.deltaTime
}

// SetDeltaTime changes delta time
func (e *coreEvent) SetDeltaTime(deltaTime uint32) {
	e.deltaTime = deltaTime
}

// EventType return 'private' event type
func (e *coreEvent) EventType() EventType {
	return e.eventType
}

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
