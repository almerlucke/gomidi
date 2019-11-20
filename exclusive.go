package midi

import (
	"errors"
	"io"
)

// SystemExclusiveEvent representation
type SystemExclusiveEvent struct {
	coreEvent
	Data []byte
}

// WriteTo writer
func (e *SystemExclusiveEvent) WriteTo(w io.Writer) (int64, error) {
	var totalBytesWritten int64

	n, err := w.Write(writeVariableLengthInteger(e.deltaTime))
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	n, err = w.Write([]byte{0xF0})
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	n, err = w.Write(writeVariableLengthInteger(uint32(len(e.Data))))
	if err != nil {
		return 0, err
	}

	totalBytesWritten += int64(n)

	n, err = w.Write(e.Data)
	if err != nil {
		return 0, err
	}

	return totalBytesWritten + int64(n), nil
}

// parseSystemExclusive parses a system exclusive event
func parseSystemExclusive(statusByte uint8, deltaTime uint32, data []byte) (event Event, bytesRead uint32, err error) {
	numBytes, bytesRead, err := readVariableLengthInteger(data)
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
		coreEvent: coreEvent{
			deltaTime: deltaTime,
			eventType: SystemExclusive,
		},
		Data: exclusiveData,
	}

	return
}
