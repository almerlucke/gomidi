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
func readVariableLengthInteger(data []byte) (result uint32, bytesRead uint32, err error) {
	foundZero := false
	err = nil

	for _, b := range data {
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

// FileHeader parses a file header from a chunk
func (c *Chunk) FileHeader() (*FileHeader, error) {
	data := c.Data
	header := &FileHeader{}

	if c.Length != 6 {
		return nil, errors.New("midi header chunk data should be 6 bytes long")
	}

	header.Format = Format(binary.BigEndian.Uint16(data))
	header.NumTracks = binary.BigEndian.Uint16(data[2:])
	header.Division = binary.BigEndian.Uint16(data[4:])

	if (header.Division >> 15) == 1 {
		header.DivisionType = DivisionFramesTicks
		header.FramesPerSecond = uint8((header.Division & 0x7FFF) >> 8)
		header.TicksPerFrame = uint8(header.Division & 0xFF)
	} else {
		header.DivisionType = DivisionTicksPerQuarterNote
		header.TicksPerQuarterNote = header.Division
	}

	return header, nil
}

// Track parses a track object from a chunk
func (c *Chunk) Track() (*Track, error) {
	data := c.Data
	runningStatusActive := false
	var runningStatusByte uint8
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

	return &Track{Events: events}, nil
}

// ReadFrom reads chunk data from reader
func (c *Chunk) ReadFrom(r io.Reader) (int64, error) {
	var totalBytes int64

	p := make([]byte, 4)
	numBytes, err := r.Read(p)
	if err != nil {
		return 0, err
	}

	totalBytes += int64(numBytes)

	c.Type = ChunkType(p)
	err = binary.Read(r, binary.BigEndian, &c.Length)
	if err != nil {
		return 0, err
	}

	c.Data = make([]byte, c.Length)
	numBytes, err = r.Read(c.Data)
	if err != nil {
		return 0, err
	}

	totalBytes += int64(numBytes)

	return totalBytes, nil
}

// ReadFrom reads a midi file from reader
func (f *File) ReadFrom(r io.Reader) (int64, error) {
	var totalBytesRead int64

	f.Chunks = []*Chunk{}
	f.Tracks = []*Track{}

	for {
		chunk := &Chunk{}
		chunkBytesRead, err := chunk.ReadFrom(r)
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, err
		}

		totalBytesRead += chunkBytesRead

		f.Chunks = append(f.Chunks, chunk)

		if chunk.Type == HeaderType {
			f.Header, err = chunk.FileHeader()
			if err != nil {
				return 0, err
			}
		} else if chunk.Type == TrackType {
			track, err := chunk.Track()
			if err != nil {
				return 0, err
			}

			f.Tracks = append(f.Tracks, track)
		}
	}

	if f.Header == nil {
		return 0, errors.New("no midi header chunk found")
	}

	return totalBytesRead, nil
}
