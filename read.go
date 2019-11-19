package midi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// parseFunction type
type parseFunction func(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error)

// Mapping from event type to parse function
var eventTypeToParseFunctionMapping = map[EventType]parseFunction{
	NoteOff:               parseNoteOff,
	NoteOn:                parseNoteOn,
	PolyphonicKeyPressure: parsePolyphonicKeyPressure,
	ControlChange:         parseControlChange,
	ProgramChange:         parseProgramChange,
	ChannelPressure:       parseChannelPressure,
	PitchWheelChange:      parsePitchWheelChange,
	SystemExclusive:       parseSystemExclusive,
	SongPositionPointer:   parseSongPositionPointer,
	SongSelect:            parseSongSelect,
	TuneRequest:           parseTuneRequest,
	TimingClock:           parseTimingClock,
	Start:                 parseStart,
	Continue:              parseContinue,
	Stop:                  parseStop,
	ActiveSensing:         parseActiveSensing,
	Meta:                  parseMeta,
}

// readVariableLengthInteger reads a variable length integer from a slice of bytes
func readVariableLengthInteger(bs []byte) (result uint32, bytesRead uint32, err error) {
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

// Events get events from chunk
func (c *Chunk) Events() ([]Event, error) {
	runningStatusActive := false
	var runningStatusByte uint8
	data := c.Data
	events := []Event{}

	for {
		deltaTime, bytesRead, err := readVariableLengthInteger(data)
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

		var parseFunc parseFunction
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
	mf.Tracks = []*Track{}

	for {
		chunk := &Chunk{}
		nb, err := chunk.ReadFrom(r)
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, err
		}

		mf.Chunks = append(mf.Chunks, chunk)

		if chunk.Type == HeaderType {
			if chunk.Length != 6 {
				return 0, errors.New("midi header chunk data should be 6 bytes long")
			}

			mf.Info = chunk.HeaderInfo()
		} else if chunk.Type == TrackType {
			events, err := chunk.Events()
			if err != nil {
				return 0, err
			}

			mf.Tracks = append(mf.Tracks, &Track{Events: events})
		}

		n += int64(nb)
	}

	if mf.Info == nil {
		return 0, errors.New("no midi header chunk found")
	}

	return n, nil
}
