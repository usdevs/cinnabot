package cinnabot

import (
	"container/heap"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// maps user arguments to a key recognised by the locations map
var aliases = map[string]string{
	"Utown":   "utown",
	"Science": "science",
	"kr":      "kr-mrt",
	"MPSH":    "mpsh",
	"Arts":    "arts",
	"yih":     "yih/engin",
	"engin":   "yih/engin",
	"Com":     "comp",
	"Biz":     "biz",
	"Cenlib":  "cenlib",
	"Law":     "law",
}

// groups of locations that should be returned together
var locations = map[string][]string{
	"utown":     {"UTOWN"},
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

//Structs for BusTiming
type BusTimes struct {
	Services []Service `json:"Services"`
}

type Service struct {
	ServiceNum string  `json:"ServiceNo"`
	Next       NextBus `json:"NextBus"`
}

type NextBus struct {
	EstimatedArrival string `json:"EstimatedArrival"`
}

//BusTimings checks the bus timings based on given location
func (cb *Cinnabot) BusTimings(msg *message) {
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
	cb.SendTextMessage(int(msg.Chat.ID), busTimingResponse(&BSH))
	return
}

func makeHeap(loc tgbotapi.Location) BusStopHeap {
	//resp, _ := http.Get("https://busrouter.sg/data/2/bus-stops.json")
	responseData, _ := ioutil.ReadFile("publicstops.json")
	points := []BusStop{}
	if err := json.Unmarshal(responseData, &points); err != nil {
		log.Print(err)
	}
	BSH := BusStopHeap{points, loc}
	heap.Init(&BSH)
	return BSH
}

//busTimingResponse returns string given a busstopheap
func busTimingResponse(BSH *BusStopHeap) string {
	returnMessage := "🤖: Here are the timings:\n\n"
	//Iteratively get data for each closest bus stop.
	for i := 0; i < 4; i++ {

		busStop := heap.Pop(BSH).(BusStop)

		busStopCode := busStop.BusStopNumber

		returnMessage += "*" + busStop.BusStopName + "*\n"

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

		bt := BusTimes{}
		if err := json.Unmarshal(responseData, &bt); err != nil {
			log.Print(err)
		}
		for j := 0; j < len(bt.Services); j++ {
			arrivalTime := bt.Services[j].Next.EstimatedArrival

			layout := "2006-01-02T15:04:05-07:00"
			t, _ := time.Parse(layout, arrivalTime)
			duration := int(t.Sub(time.Now()).Minutes())
			returnMessage += "🚍Bus " + bt.Services[j].ServiceNum + " : " + strconv.Itoa(duration+1) + " minutes\n"
		}
		returnMessage += "\n"
	}
	return returnMessage
}

//NUSBusTimes structs for unmarshalling
type Response struct {
	Result ServiceResult `json:"ShuttleServiceResult"`
}
type ServiceResult struct {
	Shuttles []Shuttle `json:"shuttles"`
}

type Shuttle struct {
	ArrivalTime     string `json:"arrivalTime"`
	NextArrivalTime string `json:"nextArrivalTime"`
	Name            string `json:"name"`
}

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
		responseString := nusBusTimingResponse(&BSH)
		cb.SendTextMessage(int(msg.Chat.ID), responseString)
		return
	}

	// Check for aliases
	code := msg.Args[0]
	alias, has_alias := aliases[code]
	if has_alias {
		code = alias
	}

	// Build response components
	responseString, ok := getLocationTimings(code)
	if !ok {
		cb.SendTextMessage(int(msg.Chat.ID), "Invalid location!")
		return
	}
	responseKeyboard := makeNUSBusKeyboard(code)

	// Send response with refresh button
	response := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: 0,
		},
		Text:                  responseString,
		ParseMode:             "Markdown",
		DisableWebPagePreview: false,
	}
	response.ReplyMarkup = responseKeyboard
	cb.SendMessage(response)
}

func (cb *Cinnabot) NUSBusResfresh(qry *callback) {
	code := qry.GetArgString()
	responseString, ok := getLocationTimings(code)
	if !ok {
		cb.SendTextMessage(int(qry.ChatID), "Something went wrong while refreshing bus timings")
		return
	}
	responseKeyboard := makeNUSBusKeyboard(code)
	// msg := tgbotapi.NewEditMessageText(qry.ChatID, qry.MsgID, responseString)
	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    qry.ChatID,
			MessageID: qry.MsgID,
		},
		Text:      responseString,
		ParseMode: "Markdown",
	}

	msg.ReplyMarkup = &responseKeyboard
	cb.SendMessage(msg)
}

func (cb *Cinnabot) NUSBusHome(qry *callback) {
	return // To be implemented if bus times is migrated to inline keyboard
}

func makeNUSBusKeyboard(code string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Refresh", "//nusbus_refresh "+code),
		),
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonData("All locations", "//nusbus_home"),
		// ),
	)
}

//makeNUSHeap returns a heap for NUS Bus timings
func makeNUSHeap(loc tgbotapi.Location) BusStopHeap {
	responseData, err := ioutil.ReadFile("nusstops.json")
	if err != nil {
		log.Print(err)
	}
	points := []BusStop{}
	if err := json.Unmarshal(responseData, &points); err != nil {
		log.Print(err)
	}
	BSH := BusStopHeap{points, loc}
	heap.Init(&BSH)
	return BSH
}

func getLocationTimings(code string) (string, bool) {
	// Get a list of bus stop codes from a location code
	responseString := ""
	locs, ok := locations[code]
	if ok {
		// Format response with timings for bus stop codes
		lines := make([]string, 0)
		lines = append(lines, "🤖: Here are the bus timings")
		for _, loc := range locs {
			lines = append(lines, getBusTimings(loc))
		}
		lines = append(lines, "Last updated: "+time.Now().Format(time.RFC822))
		responseString = strings.Join(lines, "\n\n")
	}
	return responseString, ok
}

func getBusTimings(code string) string { // for location buttons
	returnMessage := "*" + code + "*\n"
	resp, _ := http.Get("https://nextbus.comfortdelgro.com.sg/eventservice.svc/Shuttleservice?busstopname=" + code)

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}
	var bt Response
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

func nusBusTimingResponse(BSH *BusStopHeap) string { // for location-based query
	returnMessage := "🤖: Here are the bus timings\n\n"
	for i := 0; i < 3; i++ {

		stop := heap.Pop(BSH).(BusStop)

		returnMessage += "*" + stop.BusStopName + "*\n"

		resp, _ := http.Get("https://nextbus.comfortdelgro.com.sg/eventservice.svc/Shuttleservice?busstopname=" + stop.BusStopNumber)

		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}
		var bt Response
		if err := json.Unmarshal(responseData, &bt); err != nil {
			log.Print(err)
		}
		/**

				var min int
				var at string
				for j := 0; j < len(bt.Result.Shuttles); j++ {
					min = j
					for k := j; k < len(bt.Result.Shuttles); k++ {
						at = bt.Result.Shuttles[k].ArrivalTime
						log.Print(at, bt.Result.Shuttles[k].Name)

						if at == "-" {
							continue
						} else if at == "Arr" {
							min = k
							continue
						}

						val := strings.Compare(at, bt.Result.Shuttles[min].ArrivalTime)
						if val == -1 {
							log.Print("A")
							min = k
						}
					}
					bt.Result.Shuttles[j], bt.Result.Shuttles[min] = bt.Result.Shuttles[min], bt.Result.Shuttles[j]
				}
		        **/

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

		returnMessage += "\n"
	}
	return returnMessage
}

//BusStop models a public / nus bus stop
type BusStop struct {
	BusStopNumber string `json:"no"`
	Latitude      string `json:"lat"`
	Longitude     string `json:"lng"`
	BusStopName   string `json:"name"`
}

type BusStopHeap struct {
	busStopList []BusStop
	location    tgbotapi.Location
}

func (h BusStopHeap) Len() int {
	return len(h.busStopList)
}

func (h BusStopHeap) Less(i, j int) bool {
	return distanceBetween2(h.location, h.busStopList[i]) < distanceBetween2(h.location, h.busStopList[j])
}

func (h BusStopHeap) Swap(i, j int) {
	h.busStopList[i], h.busStopList[j] = h.busStopList[j], h.busStopList[i]
}

func (h *BusStopHeap) Push(x interface{}) {
	h.busStopList = append(h.busStopList, x.(BusStop))
}

func (h *BusStopHeap) Pop() interface{} {
	oldh := h.busStopList
	n := len(oldh)
	x := oldh[n-1]
	h.busStopList = oldh[0 : n-1]

	return x
}

func distanceBetween2(Loc1 tgbotapi.Location, Loc2 BusStop) float64 {

	loc2Lat, _ := strconv.ParseFloat(Loc2.Latitude, 32)
	loc2Lon, _ := strconv.ParseFloat(Loc2.Longitude, 32)

	x := math.Pow(Loc1.Latitude-loc2Lat, 2)
	y := math.Pow(Loc1.Longitude-loc2Lon, 2)
	return x + y
}
