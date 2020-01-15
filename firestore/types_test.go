package firestore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/usdevs/cinnabot/utils"
)

func TestString(t *testing.T) {
	blob := `{"stringValue":"abc"}`
	var value String

	// Unmarshalling
	if err := json.Unmarshal([]byte(blob), &value); err != nil {
		t.Errorf("JSON unmarshalling error: %s", err)
	}
	expected := "abc"
	if value.Value() != expected {
		t.Errorf("%T unmarshalling: %v expected, got %v", expected, expected, value.Value())
	}

	// Marshalling
	tmp, err := json.Marshal(value)
	if err != nil {
		t.Errorf("JSON marshalling error: %s", err)
	}
	result := string(tmp)
	if result != blob {
		t.Errorf("%T marshalling: %v expected, got %v", expected, blob, result)
	}
}

func TestInt(t *testing.T) {
	blob := `{"integerValue":"1"}`
	var value Int

	// Unmarshalling
	if err := json.Unmarshal([]byte(blob), &value); err != nil {
		t.Errorf("JSON unmarshalling error: %s", err)
	}
	expected := 1
	if value.Value() != expected {
		t.Errorf("%T unmarshalling: %v expected, got %v", expected, expected, value.Value())
	}

	// Marshalling
	tmp, err := json.Marshal(value)
	if err != nil {
		t.Errorf("JSON marshalling error: %s", err)
	}
	result := string(tmp)
	if result != blob {
		t.Errorf("%T marshalling: %v expected, got %v", expected, blob, result)
	}
}

// timeAndOffsetEqual checks if 2 times have the same value and timezone.
func timeAndOffsetEqual(t1, t2 time.Time) bool {
	timeEq := t1.Equal(t2)
	_, offset1 := t1.Zone()
	_, offset2 := t2.Zone()
	locEq := offset1 == offset2
	return timeEq && locEq
}

func TestTime(t *testing.T) {
	blob := `{"timestampValue":"2019-12-20T13:00:00Z"}`

	// Unmarshalling
	var value Time
	if err := json.Unmarshal([]byte(blob), &value); err != nil {
		t.Errorf("JSON unmarshalling error: %s", err)
	}
	loc, _ := time.LoadLocation("UTC")
	expected := time.Date(2019, time.December, 20, 13, 0, 0, 0, loc).In(utils.SgLocation())
	if !timeAndOffsetEqual(expected, value.Value()) {
		t.Errorf("%T %v expected, got %v", expected, expected, value.Value())
	}

	// Marshalling
	tmp, err := json.Marshal(value)
	if err != nil {
		t.Errorf("JSON marshalling error: %s", err)
	}
	result := string(tmp)
	if result != blob {
		t.Errorf("%T marshalling: %v expected, got %v", expected, blob, result)
	}
}

func TestBool(t *testing.T) {
	blob := `{"booleanValue":false}`
	var value Bool
	if err := json.Unmarshal([]byte(blob), &value); err != nil {
		t.Errorf("JSON unmarshalling error: %s", err)
	}
	expected := false
	if value.Value() != expected {
		t.Errorf("%T %v expected, got %v", expected, expected, value.Value())
	}

	// Marshalling
	tmp, err := json.Marshal(value)
	if err != nil {
		t.Errorf("JSON marshalling error: %s", err)
	}
	result := string(tmp)
	if result != blob {
		t.Errorf("%T marshalling: %v expected, got %v", expected, blob, result)
	}
}
