package firestore

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/usdevs/cinnabot/utils"
)

// String is a wrapper for a string in firestore, which has the JSON format {"stringValue": "abc"}.
type String string

// Time is a wrapper for a timestamp in firestore, which has the JSON format { "timestampValue": "yyyy-mm-ddThh:mm:ssZ" }.
type Time time.Time

// Int is a wrapper for an integer in firestore, which has the JSON format {"integerValue": "1"}.
type Int int

// Bool is a wrapper for a boolean in firestore, which has the JSON format {"booleanValue" : true}.
type Bool bool

type rawString struct {
	Val string `json:"stringValue"`
}

type rawTime struct {
	Val string `json:"timestampValue"`
}

type rawInt struct {
	Val string `json:"integerValue"`
}

type rawBool struct {
	Val bool `json:"booleanValue"`
}

// JSON Unmarshaller implementation

func (fs *String) UnmarshalJSON(data []byte) error {
	var raw rawString
	jsonErr := json.Unmarshal(data, &raw)
	*fs = String(raw.Val)
	return jsonErr
}

func (ft *Time) UnmarshalJSON(data []byte) error {
	var raw rawTime
	if jsonErr := json.Unmarshal(data, &raw); jsonErr != nil {
		return jsonErr
	}
	time, timeErr := time.Parse("2006-01-02T15:04:05Z", raw.Val)
	*ft = Time(time.In(utils.SgLocation())) // convert to Sg time for display purposes
	return timeErr
}

func (fi *Int) UnmarshalJSON(data []byte) error {
	var raw rawInt
	if jsonErr := json.Unmarshal(data, &raw); jsonErr != nil {
		return jsonErr
	}
	i, err := strconv.Atoi(raw.Val)
	if err != nil {
		return err
	}
	*fi = Int(i)
	return nil
}

func (fb *Bool) UnmarshalJSON(data []byte) error {
	var raw rawBool
	jsonErr := json.Unmarshal(data, &raw)
	*fb = Bool(raw.Val)
	return jsonErr
}

// JSON Marshaller implementation

func (fs String) MarshalJSON() ([]byte, error) {
	return json.Marshal(rawString{string(fs)})
}

func (ft Time) MarshalJSON() ([]byte, error) {
	timeString := time.Time(ft).UTC().Format("2006-01-02T15:04:05Z")
	return json.Marshal(rawTime{timeString})
}

func (fi Int) MarshalJSON() ([]byte, error) {
	str := strconv.Itoa(int(fi))
	return json.Marshal(rawInt{str})
}

func (fb Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(rawBool{bool(fb)})
}

// Getters

func (fs String) Value() string { return string(fs) }

func (ft Time) Value() time.Time { return time.Time(ft) }

func (fb Bool) Value() bool { return bool(fb) }

func (fi Int) Value() int { return int(fi) }

// fmt.Stringer is implemented to allow pretty printing

func (ft Time) String() string { return ft.Value().String() }
