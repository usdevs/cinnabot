// Package firestore reads from publically accessible Firestore databases using the REST API.
//
// The structs used largely mirror those used in the REST API.
// types.go contains all the types supported by this library.
package firestore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	// Operators
	LessThan           = "LESS_THAN"
	GreaterThan        = "GREATER_THAN"
	LessThanOrEqual    = LessThan + orEqual
	GreaterThanOrEqual = GreaterThan + orEqual
	Equal              = "EQUAL"

	orEqual = "_OR_EQUAL"

	baseUrl         = "https://firestore.googleapis.com/v1beta1/"
	nameFormat      = "projects/%s/databases/%s/documents" // first: project name, second: database name
	defaultDatabase = "(default)"
)

type RawDocument struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"fields"`
}

type Document struct {
	Name string      `json:"name"`
	Data interface{} `json:"fields"`
}

// ParseData parses the raw JSON data in a RawDocument to any data type using the parse function, and returns
// a Document containing the parsed data. A document containing the parsed data is returned even if the
// parse function returns an error.
func (raw RawDocument) ParseData(parse func(json.RawMessage) (interface{}, error)) (Document, error) {
	data, err := parse(raw.Data)
	if err != nil {
		log.Printf("Error in firestore/query.go while parsing JSON: %s", err)
	}
	return Document{Name: raw.Name, Data: data}, err
}

type CollectionSelector struct {
	CollectionId   string `json:"collectionId"`
	AllDescendants bool   `json:"allDescendants"`
}

type fieldString string

func (s fieldString) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Val string `json:"fieldPath"`
	}{string(s)})
}

// Filter represents a Firestore fieldFilter.
type Filter struct {
	Field          fieldString `json:"field"` // wrapper for string
	Op             string      `json:"op"`
	FirestoreValue interface{} `json:"value"` // one of the types defined in types.go
}

func (f Filter) MarshalJSON() ([]byte, error) {
	type innerFilter Filter
	return json.Marshal(struct {
		Inner innerFilter `json:"fieldFilter"`
	}{innerFilter(f)})
}

type Filters []Filter

func (fils Filters) MarshalJSON() ([]byte, error) {
	if fils == nil || len(fils) == 0 {
		result, _ := json.Marshal(struct{}{})
		return result, errors.New("Filters of a query with no filters should not be marshalled into JSON")
	} else if len(fils) == 1 {
		return json.Marshal(fils[0])
	} else {
		inner, err := json.Marshal(struct {
			Op      string   `json:"op"`
			Filters []Filter `json:"filters"`
		}{"AND", fils})
		if err != nil {
			return make([]byte, 0), err
		}
		return json.Marshal(struct {
			Inner json.RawMessage `json:"compositeFilter"`
		}{inner})
	}
}

type Query struct {
	From  []CollectionSelector `json:"from"`
	Where Filters              `json:"where"` // wrapper for []Filter
}

func (q Query) MarshalJSON() ([]byte, error) {
	var inner json.RawMessage
	var err error
	if q.Where == nil || len(q.Where) == 0 {
		type noFilter struct {
			From []CollectionSelector `json:"from"`
		}
		inner, err = json.Marshal(noFilter{q.From})
	} else {
		type tmpQuery Query // avoid infinite recursion
		inner, err = json.Marshal(tmpQuery(q))
	}
	if err != nil {
		return make([]byte, 0), err
	}
	return json.Marshal(struct {
		X json.RawMessage `json:"structuredQuery"`
	}{inner})
}

func sendRequest(projectId string, query Query) ([]byte, error) {
	url := baseUrl + fmt.Sprintf(nameFormat, projectId, defaultDatabase) + ":runQuery"
	queryBytes, queryErr := json.Marshal(query)
	if queryErr != nil {
		log.Print("Error in firestore/query.go while constructing query")
		return make([]byte, 0), queryErr
	}

	response, responseErr := http.Post(url, "application/json", bytes.NewReader(queryBytes))
	if responseErr != nil {
		log.Print("Error in firestore/query.go while receiving HTTP response")
		return make([]byte, 0), responseErr
	}
	defer response.Body.Close()

	rawData, rawDataErr := ioutil.ReadAll(response.Body)
	if rawDataErr != nil {
		log.Print("Error in firestore/query.go while reading HTTP response body")
		return rawData, rawDataErr
	}
	return rawData, nil
}

// type strTime time.Time

type queryDocument struct {
	Document RawDocument `json:"document"`
	// ReadTime strTime `json:"readTime"`
}

type errorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// parseResponse returns an error if the response format cannot be parsed, or if the response describes an error.
func parseResponse(body []byte) ([]RawDocument, error) {
	docs := make([]RawDocument, 0)

	var queryDocs []queryDocument
	if jsonErr := json.Unmarshal(body, &queryDocs); jsonErr != nil {
		log.Print("Error in firestore/query.go while parsing JSON")
		return docs, jsonErr
	}

	for _, doc := range queryDocs {
		docs = append(docs, doc.Document)
	}

	if len(docs) == 0 {
		return docs, errors.New("Invalid response format")
	}

	// Check if there are actually Documents
	if len(docs) == 1 && docs[0].Name == "" { // all firestore documents must have a name (path)
		docs = make([]RawDocument, 0)

		// Check for error response
		var errorResp []errorResponse
		json.Unmarshal(body, &errorResp)
		if len(errorResp) > 0 && errorResp[0] != (errorResponse{}) {
			log.Printf("Error response from firestore in firestore/query.go: %+v", errorResp[0])
			return docs, errors.New(fmt.Sprintf("%+v", errorResp[0]))
		}
		// otherwise this means there are no matches, so just return an empty list and no error.
	}

	return docs, nil
}

// RunQuery queries the project with the specified ID and returns the response without parsing the document data.
// An empty list of doucments is returned if any errors occur.
func RunQuery(projectId string, query Query) ([]RawDocument, error) {
	response, err := sendRequest(projectId, query)
	if err != nil {
		return make([]RawDocument, 0), err
	}
	return parseResponse(response)
}

// RunQueryAndParse queries the project with the specified ID and returns a response with the document data parsed
// using the RawDocument.ParseData function. includeErrored is used to specify whether documents which cannot
// be parsed using the parse function (ie. return an error) should be included in the returned list of documents.
func RunQueryAndParse(projectId string, query Query, parse func(json.RawMessage) (interface{}, error), includeErrored bool) ([]Document, error) {
	rawDocs, err := RunQuery(projectId, query)
	docs := make([]Document, 0, len(rawDocs))
	if err != nil {
		return docs, err
	}
	for _, raw := range rawDocs {
		doc, err := raw.ParseData(parse)
		if includeErrored || err == nil {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}
