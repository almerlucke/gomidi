package midi

import (
	"bytes"
	"encoding/binary"
	"io"
)

func writeVariableLengthValue(value uint32) []byte {
	data := []byte{}

	// Start xor with 0 byte
	xor := byte(0x0)

	for {
		// Get first 7 bits
		b := byte(value & 0x7F)

		// Xor with current xor
		b ^= xor

		// Set xor to 0x80 = 10000000 in bits
		xor = byte(0x80)

		// Push byte to front
		data = append([]byte{b}, data...)

		// Shift to next 7 bits
		value >>= 7

		// Stop if value is zero
		if value == 0 {
			break
		}
	}

	return data
}

// Chunk from track
func (t *Track) Chunk() *Chunk {
	var buf bytes.Buffer

	for _, event := range t.Events {
		event.WriteTo(&buf)
	}

	data := buf.Bytes()

	return &Chunk{
		Type:   TrackType,
		Length: uint32(len(data)),
		Data:   data,
	}
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
