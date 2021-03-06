package cinnabot

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// busStopHeap is used for GPS location based queries
type busStopHeap struct {
	busStopList []busStop
	location    tgbotapi.Location
}

func (h busStopHeap) Len() int {
	return len(h.busStopList)
}

func (h busStopHeap) Less(i, j int) bool {
	return distanceBetween2(h.location, h.busStopList[i]) < distanceBetween2(h.location, h.busStopList[j])
}

func (h busStopHeap) Swap(i, j int) {
	h.busStopList[i], h.busStopList[j] = h.busStopList[j], h.busStopList[i]
}

func (h *busStopHeap) Push(x interface{}) {
	h.busStopList = append(h.busStopList, x.(busStop))
}

func (h *busStopHeap) Pop() interface{} {
	oldh := h.busStopList
	n := len(oldh)
	x := oldh[n-1]
	h.busStopList = oldh[0 : n-1]

	return x
}

func distanceBetween2(Loc1 tgbotapi.Location, Loc2 busStop) float64 {

	loc2Lat, _ := strconv.ParseFloat(Loc2.Latitude, 32)
	loc2Lon, _ := strconv.ParseFloat(Loc2.Longitude, 32)

	x := math.Pow(Loc1.Latitude-loc2Lat, 2)
	y := math.Pow(Loc1.Longitude-loc2Lon, 2)
	return x + y
}

// busStop models a public / nus bus stop (from the JSON file)
type busStop struct {
	BusStopNumber string `json:"no"`
	Latitude      string `json:"lat"`
	Longitude     string `json:"lng"`
	BusStopName   string `json:"name"`
}

//// NUS BUSES

// structs for umarshalling timings from NUS bus API
type response struct {
	Result serviceResult `json:"ShuttleServiceResult"`
}

type serviceResult struct {
	Shuttles []shuttle `json:"shuttles"`
}

type shuttle struct {
	ArrivalTime     string `json:"arrivalTime"`
	NextArrivalTime string `json:"nextArrivalTime"`
	Name            string `json:"name"`
}

//makeNUSHeap returns a heap for NUS Bus timings
func makeNUSHeap(loc tgbotapi.Location) busStopHeap {
	responseData, err := ioutil.ReadFile("nusstops.json")
	if err != nil {
		log.Print(err)
	}
	points := []busStop{}
	if err := json.Unmarshal(responseData, &points); err != nil {
		log.Print(err)
	}
	BSH := busStopHeap{points, loc}
	heap.Init(&BSH)
	return BSH
}

// maps user arguments to a key recognised by the nusBusStopLocations map
var aliases = map[string]string{
	"kr":    "kr-mrt",
	"yih":   "yih/engin",
	"engin": "yih/engin",
	"Com":   "comp",
}

// groups of bus stops that should be returned together
var nusBusStopLocations = map[string][]string{
	"utown":     {"UTown"},
	"science":   {"S17", "LT27"},
	"kr-mrt":    {"KR-MRT", "KR-MRT-OPP"},
	"mpsh":      {"STAFFCLUB", "STAFFCLUB-OPP"},
	"arts":      {"LT13", "LT13-OPP", "AS7"},
	"yih/engin": {"YIH", "YIH-OPP", "MUSEUM", "RAFFLES"},
	"comp":      {"COM2"},
	"biz":       {"HSSML-OPP", "BIZ2", "NUSS-OPP"},
	"cenlib":    {"COMCEN", "CENLIB"},
	"law":       {"BUKITTIMAH-BTC2"},
}

// Get a list of bus stop codes from a location code (for button-based query)
func nusBusResponse_Buttons(code string) (string, bool) {
	code = strings.ToLower(code)
	responseString := ""
	locs, ok := nusBusStopLocations[code]
	if ok {
		// Format response with timings for bus stop codes
		lines := make([]string, 0)
		lines = append(lines, "🤖: Here are the bus timings")
		for _, loc := range locs {
			lines = append(lines, getNUSBusTimings(loc, loc))
		}
		lines = append(lines, "Last updated: "+time.Now().Format(time.RFC822))
		responseString = strings.Join(lines, "\n")
	}
	return responseString, ok
}

// for location-based query
func nusBusResponse_Location(BSH *busStopHeap) string {
	lines := make([]string, 0)
	lines = append(lines, "🤖: Here are the bus timings")
	for i := 0; i < 3; i++ {
		stop := heap.Pop(BSH).(busStop)
		lines = append(lines, getNUSBusTimings(stop.BusStopNumber, stop.BusStopName))
	}
	lines = append(lines, "Last updated: "+time.Now().Format(time.RFC822))
	return strings.Join(lines, "\n")
}

// fetch NUS bus timings from web API
func getNUSBusTimings(code, displayName string) string {
	returnMessage := "*" + displayName + "*\n"
	resp, _ := http.Get("https://better-nextbus.appspot.com/ShuttleService?busstopname=" + code)

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}
	var bt response
	if err := json.Unmarshal(responseData, &bt); err != nil {
		log.Print(err)
	}

	for j := 0; j < len(bt.Result.Shuttles); j++ {
		arrivalTime := bt.Result.Shuttles[j].ArrivalTime
		nextArrivalTime := bt.Result.Shuttles[j].NextArrivalTime

		if arrivalTime == "-" {
			returnMessage += "🛑" + bt.Result.Shuttles[j].Name + " : - mins\n"
			continue
		} else if arrivalTime == "1" {
			returnMessage += "🚍" + bt.Result.Shuttles[j].Name + " : " + arrivalTime + " min, " + nextArrivalTime + " mins\n"
		} else if arrivalTime == "Arr" {
			returnMessage += "🚍" + bt.Result.Shuttles[j].Name + " : " + arrivalTime + ", " + nextArrivalTime + " mins\n"
		} else {
			returnMessage += "🚍" + bt.Result.Shuttles[j].Name + " : " + arrivalTime + " mins, " + nextArrivalTime + " mins\n"
		}
	}
	return returnMessage
}

//// PUBLIC BUSES

// structs for umarshalling timings from public bus API
type busTimes struct {
	Services []service `json:"Services"`
}

type service struct {
	ServiceNum string  `json:"ServiceNo"`
	Next       nextBus `json:"NextBus"`
}

type nextBus struct {
	EstimatedArrival string `json:"EstimatedArrival"`
}

func makeHeap(loc tgbotapi.Location) busStopHeap {
	//resp, _ := http.Get("https://busrouter.sg/data/2/bus-stops.json")
	responseData, _ := ioutil.ReadFile("publicstops.json")
	points := []busStop{}
	if err := json.Unmarshal(responseData, &points); err != nil {
		log.Print(err)
	}
	BSH := busStopHeap{points, loc}
	heap.Init(&BSH)
	return BSH
}

//busTimingResponse returns string given a busstopheap
func publicBusResponse(BSH *busStopHeap) string {
	lines := make([]string, 0)
	lines = append(lines, "🤖: Here are the bus timings")
	//Iteratively get data for each closest bus stop.
	for i := 0; i < 4; i++ {

		busStop := heap.Pop(BSH).(busStop)

		busStopCode := busStop.BusStopNumber

		lines = append(lines, "*"+busStop.BusStopName+"*")

		//Send request to my transport sg for bus timing data
		client := &http.Client{}

		req, _ := http.NewRequest("GET",
			"http://datamall2.mytransport.sg/ltaodataservice/BusArrivalv2?BusStopCode="+busStopCode, nil)
		req.Header.Set("AccountKey", "l88uTu9nRjSO6VYUUwilWg==")

		resp, _ := client.Do(req)
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}

		bt := busTimes{}
		if err := json.Unmarshal(responseData, &bt); err != nil {
			log.Print(err)
		}
		for j := 0; j < len(bt.Services); j++ {
			arrivalTime := bt.Services[j].Next.EstimatedArrival

			layout := "2006-01-02T15:04:05-07:00"
			t, _ := time.Parse(layout, arrivalTime)
			duration := int(t.Sub(time.Now()).Minutes())
			s := "🚍Bus " + bt.Services[j].ServiceNum + " : " + strconv.Itoa(duration+1) + " minutes"
			lines = append(lines, s)
		}
		lines = append(lines, "")
	}
	lines = append(lines, "Last updated: "+time.Now().Format(time.RFC822))
	return strings.Join(lines, "\n")
}

//// REFRESH BUTTONS

func makeNUSBusKeyboard_Buttons(code string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Refresh", "//nusbus_refresh "+code),
		),
	)
}

func makeNUSBusKeyboard_Location(loc tgbotapi.Location) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Refresh", fmt.Sprintf("//nusbus_loc_refresh %f %f", loc.Longitude, loc.Latitude)),
		),
	)
}

func makePublicBusKeyboard(loc tgbotapi.Location) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Refresh", fmt.Sprintf("//publicbus_refresh %f %f", loc.Longitude, loc.Latitude)),
		),
	)
}

//// CINNABOT FUNCTIONS

//NUSBus retrieves the next timing for NUS Shuttle buses
func (cb *Cinnabot) NUSBus(msg *message) {
	//If no args in nusbus and arg not relevant to bus
	if len(msg.Args) == 0 || !cb.CheckArgCmdPair("/nusbus", msg.Args) {
		opt1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("UTown"), tgbotapi.NewKeyboardButton("Science"))
		opt2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Arts"), tgbotapi.NewKeyboardButton("Comp"))
		opt3 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("CenLib"), tgbotapi.NewKeyboardButton("Biz"))
		opt4 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Law"), tgbotapi.NewKeyboardButton("Yih/Engin"))
		opt5 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("MPSH"), tgbotapi.NewKeyboardButton("KR-MRT"))

		opt6B := tgbotapi.NewKeyboardButton("Here")
		opt6B.RequestLocation = true
		opt6 := tgbotapi.NewKeyboardButtonRow(opt6B)

		options := tgbotapi.NewReplyKeyboard(opt6, opt1, opt2, opt3, opt4, opt5)
		options.ResizeKeyboard = true
		options.OneTimeKeyboard = true
		options.Selective = true

		replyMsg := tgbotapi.NewMessage(int64(msg.Chat.ID), "🤖: Where are you?\n\n")
		replyMsg.ReplyMarkup = options
		cb.SendMessage(replyMsg)
		return
	}

	//Default loc: Cinnamon
	loc := &tgbotapi.Location{Latitude: 1.306671, Longitude: 103.773556}

	if msg.Location != nil {
		loc = msg.Location
		//Returns a heap of busstop data (sorted)
		BSH := makeNUSHeap(*loc)
		responseString := nusBusResponse_Location(&BSH)
		responseKeyboard := makeNUSBusKeyboard_Location(*loc)
		response := NewMessageWithButton(responseString, responseKeyboard, msg.Chat.ID)
		cb.SendMessage(response)
		return
	}

	// Check for aliases
	code := msg.Args[0]
	alias, hasAlias := aliases[code]
	if hasAlias {
		code = alias
	}

	// Build response components
	responseString, ok := nusBusResponse_Buttons(code)
	if !ok {
		cb.SendTextMessage(int(msg.Chat.ID), "Invalid location!")
		return
	}
	responseKeyboard := makeNUSBusKeyboard_Buttons(code)
	response := NewMessageWithButton(responseString, responseKeyboard, msg.Chat.ID)
	cb.SendMessage(response)
}

//BusTimings checks the public bus timings based on given location
func (cb *Cinnabot) PublicBus(msg *message) {
	if len(msg.Args) == 0 || !cb.CheckArgCmdPair("/publicbus", msg.Args) {
		opt1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Cinnamon"))
		opt2B := tgbotapi.NewKeyboardButton("Here")
		opt2B.RequestLocation = true
		opt2 := tgbotapi.NewKeyboardButtonRow(opt2B)

		options := tgbotapi.NewReplyKeyboard(opt1, opt2)

		replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "🤖: Where are you?\n\n")
		replyMsg.ReplyMarkup = options
		cb.SendMessage(replyMsg)
		return
	}
	//Asynchronous

	//Default loc: Cinnamon
	loc := &tgbotapi.Location{Latitude: 1.306671, Longitude: 103.773556}

	if msg.Location != nil {
		loc = msg.Location
	}
	//Returns a heap of busstop data (sorted)
	BSH := makeHeap(*loc)
	responseString := publicBusResponse(&BSH)
	responseKeyboard := makePublicBusKeyboard(*loc)
	response := NewMessageWithButton(responseString, responseKeyboard, msg.Chat.ID)
	cb.SendMessage(response)
}

//NUSBusHome updates the inline keyboard to a bus stop selector keyboard
/* func (cb *Cinnabot) NUSBusHome(qry *Callback) {
	return // To be implemented if bus times is migrated to inline keyboard
} */

//// REFERSH BUTTON HANDLERS

//NUSBusRefresh_Buttons handles the refresh button for messages from /nusbus -> location button
func (cb *Cinnabot) NUSBusRefresh_Buttons(qry *Callback) {
	code := qry.GetArgString()
	responseString, ok := nusBusResponse_Buttons(code)
	if !ok {
		cb.SendTextMessage(int(qry.ChatID), "Something went wrong while refreshing bus timings")
		return
	}
	responseKeyboard := makeNUSBusKeyboard_Buttons(code)
	msg := EditedMessageWithButton(responseString, responseKeyboard, qry.ChatID, qry.MsgID)
	cb.SendMessage(msg)
}

//NUSBusRefresh_Location handles the refresh button for messages from /nusbus by GPS location
func (cb *Cinnabot) NUSBusRefresh_Location(qry *Callback) {
	long, _ := strconv.ParseFloat(qry.Args[0], 64)
	lat, _ := strconv.ParseFloat(qry.Args[1], 64)

	loc := tgbotapi.Location{Longitude: long, Latitude: lat}
	heap := makeNUSHeap(loc)
	responseString := nusBusResponse_Location(&heap)
	responseKeyboard := makeNUSBusKeyboard_Location(loc)
	response := EditedMessageWithButton(responseString, responseKeyboard, qry.ChatID, qry.MsgID)
	cb.SendMessage(response)
}

//PublicBusRefresh handles the refresh button for messages from /publicbus
func (cb *Cinnabot) PublicBusRefresh(qry *Callback) {
	long, _ := strconv.ParseFloat(qry.Args[0], 64)
	lat, _ := strconv.ParseFloat(qry.Args[1], 64)

	loc := tgbotapi.Location{Longitude: long, Latitude: lat}
	BSH := makeHeap(loc)
	responseString := publicBusResponse(&BSH)
	responseKeyboard := makePublicBusKeyboard(loc)
	response := EditedMessageWithButton(responseString, responseKeyboard, qry.ChatID, qry.MsgID)
	cb.SendMessage(response)
}
