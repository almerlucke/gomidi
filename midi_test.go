package midi

import (
	"os"
	"testing"
)

func TestVariableLengthInteger(t *testing.T) {
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

// TestMidi test midi
func TestMidi(t *testing.T) {

	r, err := os.Open("data/teddybear.mid")
	if err != nil {
		t.Fatalf("err %v", err)
	}

	defer r.Close()

	mf := &File{}

	_, err = mf.ReadFrom(r)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	t.Logf("%v", *mf.Header)
	t.Logf("chunks %v\n", mf.Chunks)

	if len(mf.Tracks) > 0 {
		track := mf.Tracks[3]

		for _, event := range track.Events {
			t.Logf("deltaTime %v - event %v", event.DeltaTime(), eventTypeToString(event.EventType()))
		}
	}
}
