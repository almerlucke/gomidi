package midi

import (
	"os"
	"testing"
)

func TestReadVariableLengthInteger(t *testing.T) {
	// Test valid input
	bs := make([]byte, 2)
	bs[0] = 0xFF
	bs[1] = 0x7F

	v, n, err := readVariableLengthInteger(bs)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if v != 16383 {
		t.Errorf("expected value to be 16383, value returned is %v", v)
	}

	if n != 2 {
		t.Errorf("expected num bytes read to be 2, num bytes returned is %v", n)
	}

	bs = make([]byte, 2)
	bs[0] = 0x87
	bs[1] = 0x68

	v, n, err = readVariableLengthInteger(bs)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if v != 1000 {
		t.Errorf("expected value to be 1000, value returned is %v", v)
	}

	if n != 2 {
		t.Errorf("expected num bytes read to be 2, num bytes returned is %v", n)
	}

	bs = make([]byte, 3)
	bs[0] = 0xBD
	bs[1] = 0x84
	bs[2] = 0x40

	v, n, err = readVariableLengthInteger(bs)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if v != 1000000 {
		t.Errorf("expected value to be 1000000, value returned is %v", v)
	}

	if n != 3 {
		t.Errorf("expected num bytes read to be 3, num bytes returned is %v", n)
	}

	// Test invalid input
	bs = make([]byte, 2)
	bs[0] = 0xFF
	bs[1] = 0xFF

	v, n, err = readVariableLengthInteger(bs)
	if err == nil {
		t.Errorf("expected ReadVariableLengthInteger to return an error")
	}
}

func TestWriteVariableLengthVariable(t *testing.T) {
	bs := make([]byte, 1)
	bs[0] = 0x0

	data := writeVariableLengthValue(0)
	if len(data) != len(bs) {
		t.Fatalf("0: inequal length of bytes %d - %d", len(data), len(bs))
	}

	for index, b := range data {
		if bs[index] != b {
			t.Errorf("0: byte %d is not equal", index)
		}
	}

	t.Log("0 passed")

	bs = make([]byte, 2)
	bs[0] = 0xFF
	bs[1] = 0x7F

	data = writeVariableLengthValue(16383)
	if len(data) != len(bs) {
		t.Fatalf("16383: inequal length of bytes %d - %d", len(data), len(bs))
	}

	for index, b := range data {
		if bs[index] != b {
			t.Errorf("16383: byte %d is not equal", index)
		}
	}

	t.Log("16383 passed")

	bs = make([]byte, 2)
	bs[0] = 0x87
	bs[1] = 0x68

	data = writeVariableLengthValue(1000)
	if len(data) != len(bs) {
		t.Fatalf("1000: inequal length of bytes %d - %d", len(data), len(bs))
	}

	for index, b := range data {
		if bs[index] != b {
			t.Errorf("1000: byte %d is not equal", index)
		}
	}

	t.Log("1000 passed")

	bs = make([]byte, 3)
	bs[0] = 0xBD
	bs[1] = 0x84
	bs[2] = 0x40

	data = writeVariableLengthValue(1000000)

	if len(data) != len(bs) {
		t.Fatalf("1000000: inequal length of bytes %d - %d", len(data), len(bs))
	}

	for index, b := range data {
		if bs[index] != b {
			t.Errorf("1000000: byte %d is not equal", index)
		}
	}

	t.Log("1000000 passed")

	data = writeVariableLengthValue(1152)
	value, _, _ := readVariableLengthInteger(data)
	t.Logf("returned value %v", value)
}

// TestMidi test midi
func TestMidi(t *testing.T) {

	fo, err := os.Open("data/teddybear_test.mid")
	if err != nil {
		t.Fatalf("err %v", err)
	}

	defer fo.Close()

	mf := &File{}

	_, err = mf.ReadFrom(fo)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	t.Logf("%v", *mf.Header)
	t.Logf("chunks %v\n", mf.Chunks)

	if len(mf.Tracks) > 0 {
		track := mf.Tracks[1]

		for index, event := range track.Events {
			if index < 10 {
				t.Logf("%v", event)
			}
		}
	}

	// mft := &File{}
	// mft.Chunks = []*Chunk{mf.Chunks[0]}
	// for _, track := range mf.Tracks {
	// 	mft.Chunks = append(mft.Chunks, track.Chunk())
	// }

	// f, err := os.Create("data/teddybear_test.mid")
	// defer f.Close()
	// mft.WriteTo(f)
}
