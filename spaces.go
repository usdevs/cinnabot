package cinnabot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const GREATER_THAN = "GREATER_THAN" //for firestore queries (use with makeQuery)

// PARSING AND QUERYING:

// makeQuery returns a string to be used as the request body of a firestore query
// jsonValue must be either a jsonTime or jsonString
// see https://developers.google.com/apis-explorer/#search/firestore/firestore/v1beta1/firestore.projects.databases.documents.runQuery and https://cloud.google.com/firestore/docs/query-data/queries for details
func makeQuery(fieldName string, op string, jsonValue interface{}) string {
	formatString := `
	{
	"structuredQuery": {
		"where": {
			"fieldFilter": {
				"field": { "fieldPath": "%s" },
				"op": "%s" ,
				"value": %s
			}
		},
		"from": [ {"collectionId": "events"} ]
	}
	}
	`

	valueAsJson, _ := json.Marshal(jsonValue)
	return fmt.Sprintf(formatString, fieldName, op, string(valueAsJson))
}

// A query will return a list of objects like the one below (some fields omitted for berevity).
// See https://developers.google.com/apis-explorer/#search/firestore/firestore/v1beta1/firestore.projects.databases.documents.runQuery
/*
{"document": {
    "name": "projects/usc-website-206715/databases/(default)/documents/events/FEHpJv0xavSNCEhqE7MS",
    "fields": {
      "name": {"stringValue": "Mass Check-In"},
      "venueName": {"stringValue": "Chatterbox" },
      "fullDay": {"booleanValue": true },
      "venue": {"stringValue": "V4TTAgG9fe4bjSm4Vl3M"},
      "endDate": {"timestampValue": "2019-01-16T15:30:00Z"},
      "startDate": {"timestampValue": "2019-01-15T16:00:00Z"}
    },
    "createTime": "2018-11-08T09:54:34.250175Z",
    "updateTime": "2018-11-08T09:54:34.250175Z"
  },
  "readTime": "2018-12-19T08:47:02.232588Z"}
*/

// jsonString is a wrapper for string values returned by firestore/used in firestore queries which use the format {"stringValue" : "acutal string"}
type jsonString string

// jsonTime is a wrapper for timestamps returned from firestore/ used in firestore queries which use the format {"timestampValue": yyyy-mm-ddThh:mm:ssZ}
type jsonTime time.Time

type rawEventFields struct {
	Name  jsonString `json:"name"`
	Venue jsonString `json:"venueName"`
	Start jsonTime   `json:"startDate"`
	End   jsonTime   `json:"endDate"`
}

type rawEvent struct {
	Fields rawEventFields `json:"fields"`
}

// queryDocument represents 1 firestore document (for 1 event). A query returns []eventsQueryDocument
type queryDocument struct {
	RawEvent rawEvent `json:"document"`
}

func (js *jsonString) UnmarshalJSON(data []byte) error {
	var rawString struct {
		Val string `json:"stringValue"`
	}

	jsonErr := json.Unmarshal(data, &rawString)
	*js = jsonString(rawString.Val)
	return jsonErr
}

func (js jsonString) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Val string `json:"stringValue"`
	}{string(js)})
}

func (jt *jsonTime) UnmarshalJSON(data []byte) error {
	var rawString struct {
		Val string `json:"timestampValue"`
	}
	if jsonErr := json.Unmarshal(data, &rawString); jsonErr != nil {
		return jsonErr
	}
	time, timeErr := time.Parse("2006-01-02T15:04:05Z", rawString.Val)

	*jt = jsonTime(time)
	return timeErr
}

func (jt jsonTime) MarshalJSON() ([]byte, error) {
	timeString := time.Time(jt).Format("2006-01-02T15:04:05Z")
	return json.Marshal(struct {
		Val string `json:"timestampValue"`
	}{timeString})
}

// event converts rawEventFields into an Event
func (jef rawEventFields) event() Event {
	return Event{string(jef.Name), string(jef.Venue), time.Time(jef.Start), time.Time(jef.End)}
}

// SORTING AND FILTERING:

// Event represents a booking
// to add other fields, modify: Event, rawEventFields and rawEventFields.event()
type Event struct {
	Name  string
	Venue string
	Start time.Time
	End   time.Time
}

// Space represents a list of Events at the same venue
type Space []Event

// Spaces is a list of spaces
type Spaces []Space

// byStartDate implements sort.Interface for a Events in a Space based on space[i].Start field
type byStartDate Space

// used with filter functions
type eventPredicate func(Event) bool

// getSpaces returns the Events (as Spaces) from firestore which match the query (made using makeQuery)
func getSpaces(query string) Spaces {

	url := "https://firestore.googleapis.com/v1beta1/projects/usc-website-206715/databases/(default)/documents:runQuery?fields=document" //do not return createTime,updateTime and readTime

	spaces := make(Spaces, 0)

	response, responseErr := http.Post(url, "application/json", strings.NewReader(query))
	if responseErr != nil {
		log.Print("Error in spaces.go while receiving HTTP response")
		return spaces
	}
	defer response.Body.Close()

	rawData, rawDataErr := ioutil.ReadAll(response.Body)
	if rawDataErr != nil {
		log.Print("Error in spaces.go while reading HTTP response body")
		return spaces
	}

	var queryDocs []queryDocument
	if jsonErr := json.Unmarshal(rawData, &queryDocs); jsonErr != nil {
		log.Print("error in spaces.go while parsing JSON")
		return spaces
	}

	for _, doc := range queryDocs {
		spaces.addEvent(doc.RawEvent.Fields.event())
	}

	return spaces
}

// getSpacesAfter returns the Events (as Spaces) whose endDate is after the specified date
func getSpacesAfter(date time.Time) Spaces {
	query := makeQuery("endDate", GREATER_THAN, jsonTime(date.UTC())) //convert to UTC as firestore stores UTC datetimes
	return getSpaces(query)
}

// getName returns the name of the Space (ie. venue). It is assumed all events have been correctly added to the Space.
// If a Space has no Events, then an empty string is returned.
func (space Space) getName() string {
	if len(space) > 0 {
		return space[0].Venue
	}

	return ""

}

// addEvent adds an Event to a Space. Does not check whether the event to be added has the same venue as the other events in that Space.
func (space *Space) addEvent(event Event) {
	*space = append(*space, event)
}

// addEvent adds an Event to the appropriate Space, if it exists. Otherwise, a new Space is created.
func (spaces *Spaces) addEvent(event Event) {
	for i, space := range *spaces {
		if space.getName() == event.Venue {
			(*spaces)[i].addEvent(event)
			return
		}
	}

	// space not created yet
	// create new space
	space := make([]Event, 1)
	space[0] = event
	*spaces = append(*spaces, space)
}

func (events byStartDate) Len() int           { return len(events) }
func (events byStartDate) Swap(i, j int)      { events[i], events[j] = events[j], events[i] }
func (events byStartDate) Less(i, j int) bool { return events[i].Start.Before(events[j].Start) }

// sortByStartDate returns a new Space sorted in ascending order by Event.Start
func (space Space) sortByStartDate() Space {
	sort.Sort(byStartDate(space))
	return space
}

func (spaces Spaces) sortByStartDate() Spaces {
	sortedSpaces := make(Spaces, len(spaces))
	for i, space := range spaces {
		sortedSpaces[i] = space.sortByStartDate()
	}
	return sortedSpaces
}

// filter returns a new Space with only Events which satisfy the predicate.
func (space Space) filter(predicate eventPredicate) Space {
	filteredEvents := make(Space, 0, len(space))
	for _, event := range space {
		if predicate(event) {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents
}

// filter returns a new Spaces with only Events which satisfy the predicate
func (spaces Spaces) filter(predicate eventPredicate) Spaces {
	filteredSpaces := make(Spaces, 0, len(spaces))
	for _, space := range spaces {
		filteredSpace := space.filter(predicate)
		//do not add spaces with no events
		if len(filteredSpace) > 0 {
			filteredSpaces = append(filteredSpaces, filteredSpace)
		}
	}
	return filteredSpaces
}

// eventBetween returns true if an event occurs between the 2 times given. assumes firstDate and lastDate are not equal
func eventBetween(firstDate, lastDate time.Time) eventPredicate {
	return func(e Event) bool {
		if e.Start.Before(firstDate) {
			return e.End.After(firstDate)
		}
		return e.Start.Before(lastDate)
	}
}

func eventDuring(date time.Time) eventPredicate {
	return func(e Event) bool {
		return (e.Start.Before(date) && e.End.After(date)) || e.Start.Equal(date)
	}
}

func eventOnDay(date time.Time) eventPredicate {
	return eventBetween(startOfDay(date), endOfDay(date))
}

// eventBetweenDays returns true if an event occurs between the start of the first day and the end of the last day
func eventBetweenDays(firstDate, lastDate time.Time) eventPredicate {
	return eventBetween(startOfDay(firstDate), endOfDay(lastDate))
}

// startOfDay returns a new time.Time with the time set to 00:00 (SG time)
func startOfDay(date time.Time) time.Time {
	localDate := date.Local()
	return time.Date(localDate.Year(), localDate.Month(), localDate.Day(), 0, 0, 0, 0, localDate.Location())
}

// endOfDay returns the next day 00:00
func endOfDay(date time.Time) time.Time {
	return startOfDay(date.AddDate(0, 0, 1))
}

// DISPLAYING:

// FormatDate formats a time.Time into date in a standardised format
func FormatDate(t time.Time) string {
	return t.Local().Format("Mon 02 Jan 06")
}

// FormatTime formats a time.Time into a time in a standardised format
func FormatTime(t time.Time) string {
	localT := t.Local()
	if localT.Minute() == 0 {
		return localT.Format("03PM")
	}
	return localT.Format("03:04PM")
}

// FormatTimeDate formats a time.Time into a full time and date, in a standardised format
func FormatTimeDate(t time.Time) string {
	return fmt.Sprintf("%s, %s", FormatTime(t), FormatDate(t))
}

// timeInfo returns the time info of an Event in a readable format with duplicate info minimised
func (event *Event) timeInfo() string {
	start := event.Start
	end := event.End
	y1, m1, d1 := start.Local().Date()
	y2, m2, d2 := end.Local().Date()

	if y1 == y2 && m1 == m2 && d1 == d2 {
		return fmt.Sprintf("%s to %s, %s", FormatTime(start), FormatTime(end), FormatDate(start))
	}
	return fmt.Sprintf("%s to %s", FormatTimeDate(start), FormatTimeDate(end))
}

// toString returns a string with the name and date/time information of an Event, with event name bolded.
func (event Event) toString() string {
	return fmt.Sprintf("*%s:* %s", event.Name, event.timeInfo())
}

// toString returns a string with the name of the space, followed by a list of the Events occuring. Only the name of the space will be printed if the space contains no Events.
func (space Space) toString() string {
	sortedSpaces := space.sortByStartDate()
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=======================\n%s\n=======================\n", space.getName()))
	for _, event := range sortedSpaces {
		sb.WriteString(event.toString())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// toString returns a string with a list of spaces (using Space.toString). If Spaces is empty [No bookings recorded] will be returned instead.
func (spaces Spaces) toString() string {

	if len(spaces) == 0 {
		return "[No bookings recorded]"
	}

	var sb strings.Builder
	for _, space := range spaces {
		sb.WriteString(space.toString())
	}
	return sb.String()
}

// bookingsNowMessage returns currently ongoing events
func bookingsNowMessage() string {
	spaces := getSpacesAfter(time.Now())
	message := fmt.Sprintf("Displaying bookings ongoing right now (%s):\n\n", FormatTimeDate(time.Now()))
	message += spaces.filter(eventDuring(time.Now())).toString()
	return message
}

// bookingsTodayMessage returns events which are happening today. Excludes events which have already finished.
func bookingsTodayMessage() string {
	spaces := getSpacesAfter(time.Now())
	message := "Displaying bookings for today:\n\n"
	message += spaces.filter(eventOnDay(time.Now())).toString()
	return message
}

// bookingsComingWeekMessage returns events which will happen/are happening in the next 7 days. Excludes events which have already finished.
func bookingsComingWeekMessage() string {
	now := time.Now()
	weekLater := now.AddDate(0, 0, 7)
	spaces := getSpacesAfter(time.Now())
	message := fmt.Sprintf("Displaying bookings 7 days from now (%s to %s):\n\n", FormatDate(now), FormatDate(weekLater))
	message += spaces.filter(eventBetweenDays(now, weekLater)).toString()
	return message
}

// bookingsBetweenMessage returns events which will occur between the specified dates. May include events which have already finished as of now.
func bookingsBetweenMessage(firstDate, lastDate time.Time) string {
	spaces := getSpacesAfter(startOfDay(firstDate))
	message := fmt.Sprintf("Displaying bookings from %s to %s:\n\n", FormatDate(firstDate), FormatDate(lastDate))
	message += spaces.filter(eventBetweenDays(firstDate, lastDate)).toString()
	return message
}

// bookingsOnDateMessage returns events which will occur on the specified date. May include events which have already finished as of now.
func bookingsOnDateMessage(date time.Time) string {
	spaces := getSpacesAfter(startOfDay(date))
	message := fmt.Sprintf("Displaying all bookings on %s:\n\n", FormatDate(date))
	message += spaces.filter(eventOnDay(date)).toString()
	return message
}

// ParseDDMMYYDate parses user-inputted dd/mm/yy date into time.Time
func ParseDDMMYYDate(date string) (time.Time, error) {
	//Attempt to parse as dd/mm/yy
	format := "02/01/06"
	t, err := time.Parse(format, date)

	if err != nil {
		// Attempt to parse as dd/m/yy
		format = "02/1/06"
		t, err = time.Parse(format, date)
	}
	if err != nil {
		// Attempt to parse as d/mm/yy
		format = "2/01/06"
		t, err = time.Parse(format, date)
	}
	if err != nil {
		// Attempt to parse as d/m/yy
		format = "2/1/06"
		t, err = time.Parse(format, date)
	}
	if err != nil {
		// Attempt to parse as some form of dd/mm
		// Attempt to parse as dd/mm
		format = "02/01"
		t, err = time.Parse(format, date)
		if err != nil {
			// Attempt to parse as dd/m
			format = "02/1"
			t, err = time.Parse(format, date)
		}
		if err != nil {
			// Attempt to parse as d/mm
			format = "2/01"
			t, err = time.Parse(format, date)
		}
		if err != nil {
			// Attempt to parse as d/m
			format = "2/1"
			t, err = time.Parse(format, date)
		}

		// Check if one of the dd/mm checks have worked
		if err == nil {
			// return t, but using the current year
			year := time.Now().Year()
			t = time.Date(year, t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
	}
	return t, err
}

//Spaces is the primary Cinnabot Spaces method that the end user interacts with.
//	"/spaces" displays bookings for today.
//	"/spaces now" displays bookings right this moment.
//	"/spaces week" displays bookings in the next 7 days.
//	"/spaces dd/mm/yy" displays bookings on the given date.
//	"/spaces dd/mm/yy dd/mm/yy" displays bookings in the given interval (limited to one month++).
//	"/spaces help" informs the user of available commands.
//
//	Extra arguments are ignored.
//	Unparseable commands return the help menu.
func (cb *Cinnabot) Spaces(msg *message) {
	cb.SendTextMessage(msg.From.ID, spacesMsg(msg))
}

// for easier debugging
func spacesMsg(msg *message) string {
	toSend := ""

	if len(msg.Args) == 0 || msg.Args[0] == "today" {
		toSend += bookingsTodayMessage()
	} else if msg.Args[0] == "now" {
		toSend += bookingsNowMessage()
	} else if msg.Args[0] == "week" {
		toSend += bookingsComingWeekMessage()
	} else if msg.Args[0] == "tomorrow" {
		toSend += bookingsOnDateMessage(time.Now().AddDate(0, 0, 1))
	} else {
		//try to parse date
		t0, err0 := ParseDDMMYYDate(msg.Args[0])

		if err0 == nil {
			// First argument is a valid date
			// Attempt to parse second argument, if exists, and show BookingsBetween(t0, t1)
			if len(msg.Args) >= 2 {
				t1, err1 := ParseDDMMYYDate(msg.Args[1])
				if err1 == nil {
					// Check if the interval is too long
					if t0.AddDate(0, 0, 33).Before(t1) {
						toSend += "The time interval is too long. Please restrict it to at most one month."
					} else {
						toSend += bookingsBetweenMessage(t0, t1)
					}
				}
			} else {
				toSend += bookingsOnDateMessage(t0)
			}
		}
	}

	if toSend == "" {
		// i.e., if arguments could not be parsed as above
		if msg.Args[0] != "help" {
			toSend += "Cinnabot was unable to understand your command.\n\n"
		}

		toSend += "To use the '/spaces' command, type one of the following:\n'/spaces' : to view all bookings for today\n'/spaces now' : to view bookings active at this very moment\n'/spaces week' : to view all bookings for this week\n'/spaces dd/mm(/yy)' : to view all bookings on a specific day\n'/spaces dd/mm(/yy) dd/mm(/yy)' : to view all bookings in a specific range of dates"
	}

	return toSend
}
