package midi

import (
	"encoding/binary"
	"errors"
	"io"
)

// ChunkType indicates the type of chunk we are dealing with
type ChunkType string

const (
	// HeaderType indicates a midi header chunk
	HeaderType ChunkType = "MThd"
	// TrackType indicates a midi track chunk
	TrackType ChunkType = "MTrk"
)

// Format type
type Format uint16

const (
	// Format0 midi file
	Format0 Format = 0
	// Format1 midi file
	Format1 Format = 1
	// Format2 midi file
	Format2 Format = 2
)

// DivisionType midi delta time
type DivisionType uint8

const (
	// DivisionTicksPerQuarterNote division
	DivisionTicksPerQuarterNote DivisionType = 0
	// DivisionFramesTicks division
	DivisionFramesTicks DivisionType = 1
)

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

type Track struct {
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
