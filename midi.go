package midi

import (
	"io"
)

// ChunkType indicates the type of chunk we are dealing with
type ChunkType string

// Format type
type Format uint16

// DivisionType midi delta time
type DivisionType uint8

// EventType to identify midi events
type EventType uint8

// Chunk holds midi chunk information
type Chunk struct {
	io.WriterTo
	io.ReaderFrom
	Type   ChunkType
	Length uint32
	Data   []byte
}

// FileHeader holds midi file header info
type FileHeader struct {
	Format              Format
	NumTracks           uint16
	Division            uint16
	DivisionType        DivisionType
	TicksPerQuarterNote uint16
	FramesPerSecond     uint8
	TicksPerFrame       uint8
}

// Track holds track info (events)
type Track struct {
	Events []Event
}

// File holds all midi chunks and other info
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

// Event interface for all midi events
type Event interface {
	DeltaTime() uint32
	EventType() EventType
}

// coreEvent to include by other event structs to satisfy Event interface
type coreEvent struct {
	deltaTime uint32
	eventType EventType
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
