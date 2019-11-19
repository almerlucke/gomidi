package midi

import (
	"encoding/binary"
	"io"
)

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
