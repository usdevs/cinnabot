package firestore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/usdevs/cinnabot/utils"
)

func TestQuery(t *testing.T) {
	// 1 filter
	query := Query{
		From:  []CollectionSelector{{"coll1", false}},
		Where: []Filter{{Field: "abc", Op: GreaterThanOrEqual, FirestoreValue: Int(1)}},
	}
	result, err := json.Marshal(query)
	if err != nil {
		t.Error(err)
	}
	// t.Log(string(result))
	expected := `{"structuredQuery":{"from":[{"collectionId":"coll1","allDescendants":false}],"where":{"fieldFilter":{"field":{"fieldPath":"abc"},"op":"GREATER_THAN_OR_EQUAL","value":{"integerValue":"1"}}}}}`
	if string(result) != expected {
		t.Errorf("Error in marshalling query, expected %s, got %s", expected, string(result))
	}

	// no filters
	query = Query{
		From:  []CollectionSelector{{"coll1", false}},
		Where: make(Filters, 0),
	}
	result, err = json.Marshal(query)
	if err != nil {
		t.Error(err)
	}
	// t.Log(string(result))
	expected = `{"structuredQuery":{"from":[{"collectionId":"coll1","allDescendants":false}]}}`
	if string(result) != expected {
		t.Errorf("Error in marshalling query with no filter, expected %s, got %s", expected, string(result))
	}

	// 2 filters
	query = Query{
		From:  []CollectionSelector{{"coll1", false}},
		Where: []Filter{{Field: "abc", Op: GreaterThanOrEqual, FirestoreValue: Int(1)}, {Field: "xyz", Op: GreaterThanOrEqual, FirestoreValue: Int(2)}},
	}
	result, err = json.Marshal(query)
	if err != nil {
		t.Error(err)
	}
	// t.Log(string(result))
	expected = `{"structuredQuery":{"from":[{"collectionId":"coll1","allDescendants":false}],"where":{"compositeFilter":{"op":"AND","filters":[{"fieldFilter":{"field":{"fieldPath":"abc"},"op":"GREATER_THAN_OR_EQUAL","value":{"integerValue":"1"}}},{"fieldFilter":{"field":{"fieldPath":"xyz"},"op":"GREATER_THAN_OR_EQUAL","value":{"integerValue":"2"}}}]}}}}`
	if string(result) != expected {
		t.Errorf("Error in marshalling query with multiple filters, expected %s, got %s", expected, string(result))
	}
}

type testData struct {
	PinNo       Int    `json:"pinNo"`
	Name        String `json:"name"`
	On          Bool   `json:"on"`
	TimeChanged Time   `json:"timeChanged"` // use omitempty as needed
}

func (t testData) equal(other testData) bool {
	return t.PinNo == other.PinNo && t.Name == other.Name && t.On == other.On && timeAndOffsetEqual(t.TimeChanged.Value(), other.TimeChanged.Value())
}

func TestDocument(t *testing.T) {
	blob := `
	[{"document": {
		"name": "projects/usc-website-206715/databases/(default)/documents/events/FEHpJv0xavSNCEhqE7MS",
		"fields": {
			"pinNo": {"integerValue": "1" },
			"on": {"booleanValue": true},
			"name": {"stringValue": "A"},
			"timeChanged": {"timestampValue": "2019-12-20T15:00:00Z"}
		},
		"createTime": "2018-11-08T09:54:34.250175Z",
		"updateTime": "2018-11-08T09:54:34.250175Z"
	},
	"readTime": "2018-12-19T08:47:02.232588Z"}]`

	loc, _ := time.LoadLocation("UTC")
	expected := testData{1, "A", true, Time(time.Date(2019, time.December, 20, 15, 0, 0, 0, loc).In(utils.SgLocation()))}

	rawdocs, err := parseResponse([]byte(blob))
	if err != nil {
		t.Errorf("Error in parseRespones: %s", err)
	}
	t.Log(string(rawdocs[0].Data))
	parse := func(data json.RawMessage) (interface{}, error) {
		obj := testData{}
		err := json.Unmarshal(data, &obj)
		return obj, err
	}
	doc, err := rawdocs[0].ParseData(parse)
	if err != nil {
		t.Errorf("Error in parsing raw document: %s", err)
	}
	t.Logf("%+v", doc)
	if !expected.equal(doc.Data.(testData)) {
		t.Errorf("Parsed data is incorrect, expected %+v, got %+v", expected, doc.Data)
	}
	// Verify that type conversion works
	// typed := doc.Data.(testData)
	// t.Log(typed)
}

// func TestSendQuery(t *testing.T) {
// 	query := Query{
// 		From:  []CollectionSelector{CollectionSelector{"laundry_status", false}},
// 		Where: []Filter{Filter{Field: "abc", Op: Equal, FirestoreValue: Int(1)}, Filter{Field: "xyz", Op: GreaterThanOrEqual, FirestoreValue: Int(2)}},
// 	}

// 	docs, err := RunQuery("usc-laundry-test", query)
// 	t.Logf("%+v", docs)
// 	t.Log(err)

// 	// 	resp, err := sendRequest("usc-laundry-test", query)
// 	// 	t.Log(string(resp))
// 	// 	// no error returned if not found/invalid query. Only error if invliad URL/offline.
// 	// 	// [{"readTime": "2019-12-23T04:31:05.621387Z"}] returned if no matches

// 	// 	if err != nil || resp == nil || len(resp) == 0 {
// 	// 		t.Error(err)
// 	// 	}

// 	// 	docs, parseErr := parseResponse(resp)
// 	// 	t.Logf("%+v", docs)
// 	// 	t.Log(parseErr)
// }
