package cinnabot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"

	"net/http"
	"sort"
	"strconv"
	"strings"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/usdevs/cinnabot/model"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

//Test functions [Not meant to be used in bot]
// SayHello says hi.
func (cb *Cinnabot) SayHello(msg *message) {
	cb.SendTextMessage(int(msg.Chat.ID), "Hello there, "+msg.From.FirstName+"!")
}

// Echo parrots back the argument given by the user.
func (cb *Cinnabot) Echo(msg *message) {
	if len(msg.Args) == 0 {
		replyMsg := tgbotapi.NewMessage(int64(msg.Message.From.ID), "/echo Cinnabot Parrot Mode 🤖\nWhat do you want me to parrot?\n\n")
		replyMsg.BaseChat.ReplyToMessageID = msg.MessageID
		replyMsg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		cb.SendMessage(replyMsg)
		return
	}
	response := "🤖: " + strings.Join(msg.Args, " ")
	cb.SendTextMessage(int(msg.Chat.ID), response)
}

// Capitalize returns a capitalized form of the input string.
func (cb *Cinnabot) Capitalize(msg *message) {
	cb.SendTextMessage(int(msg.Chat.ID), strings.ToUpper(strings.Join(msg.Args, " ")))
}

//Start initializes the bot
func (cb *Cinnabot) Start(msg *message) {
	text := "Hello there " + msg.From.FirstName + "!\n\n" +
		"Im Cinnabot🤖. I am made by my owners to serve the residents of Cinnamon college!\n" +
		"Im always here to /help if you need it!"

	cb.SendTextMessage(int(msg.Chat.ID), text)
}

// Help gives a list of handles that the user may call along with a description of them
func (cb *Cinnabot) Help(msg *message) {
	if len(msg.Args) > 0 {

		if msg.Args[0] == "spaces" {
			text :=
				"To use the '/spaces' command, type one of the following:\n" +
					"'/spaces' : to view all bookings for today\n'/spaces now' : to view bookings active at this very moment\n" +
					"'/spaces week' : to view all bookings for this week\n'/spaces dd/mm(/yy)' : to view all bookings on a specific day\n" +
					"'/spaces dd/mm(/yy) dd/mm(/yy)' : to view all bookings in a specific range of dates"
			cb.SendTextMessage(int(msg.Chat.ID), text)
			return
		} else if msg.Args[0] == "resources" {
			text :=
				"/resources <tag>: searches resources for a specific tag\n" +
					"/resources: returns all tags"
			cb.SendTextMessage(int(msg.Chat.ID), text)
			return
		} else if msg.Args[0] == "publicbus" {
			text :=
				"/publicbus : publicbus\n" +
					"Sending your location (ignore the buttons) after running the above command will allow to get bus timings for bus stops around any location."
			cb.SendTextMessage(int(msg.Chat.ID), text)
			return
		}
	}
	text :=
		"Here are a list of functions to get you started 🤸 \n" +
			"/about: to find out more about me\n" +
			"/publicbus: public bus timings for bus stops around your location\n" +
			"/nusbus: nus bus timings for bus stops around your location\n" +
			"/weather: 2h weather forecast\n" +
			"/resources: list of important resources!\n" +
			"/spaces: list of space bookings\n" +
			"/feedback: to give feedback\n\n" +
			"_*My creator actually snuck in a few more functions🕺 *_\n" +
			"Try using /help <func name> to see what I can _really_ do"
	cb.SendTextMessage(int(msg.Chat.ID), text)
}

// About returns a link to Cinnabot's source code.
func (cb *Cinnabot) About(msg *message) {
	cb.SendTextMessage(int(msg.Chat.ID), "Touch me: https://github.com/usdevs/cinnabot")
}

//Link returns useful resources
func (cb *Cinnabot) Resources(msg *message) {

	//If no args in resources and arg not relevant
	if len(msg.Args) == 0 || !cb.CheckArgCmdPair("/resources", msg.Args) {
		opt1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Telegram"), tgbotapi.NewKeyboardButton("Links"))
		opt2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Interest Groups"), tgbotapi.NewKeyboardButton("Everything"))

		options := tgbotapi.NewReplyKeyboard(opt1, opt2)

		replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "🤖: How can I help you?\n\n")
		replyMsg.ReplyMarkup = options
		cb.SendMessage(replyMsg)
		return
	}

	robotSays := "🤖: Here you go!\n\n"

	switch msg.Args[0] {
	case "telegram", "links", "interest":
		cb.SendTextMessage(int(msg.Chat.ID), robotSays+getResources(msg.Args[0]))
	case "everything":
		cb.SendTextMessage(int(msg.Chat.ID), robotSays+getResources("telegram")+"\n\n"+getResources("links")+"\n\n"+getResources("interest"))
	}

	return
	/*	var key string = strings.ToLower(strings.Join(msg.Args, " "))
		log.Print(key)
		_, ok := resources[key]
		if ok {
			cb.SendTextMessage(int(msg.Chat.ID), resources[key])
		} else {
			var values string = ""
			for key, _ := range resources {
				values += key + " : " + resources[key] + "\n"
			}
			msg := tgbotapi.NewMessage(msg.Chat.ID, values)
			msg.DisableWebPagePreview = true
			msg.ParseMode = "markdown"
			cb.SendMessage(msg)
		} */
}

type resources struct {
	Telegram       map[string]string `json:"telegram"`
	Links          map[string]string `json:"links"`
	InterestGroups map[string]string `json:"interest_groups"`
}

func getResources(code string) string { // for resources buttons
	var (
		resources resources
		jsonBytes []byte
		err       error
	)

	jsonBytes, err = ioutil.ReadFile("../resources.json")
	if err != nil {
		fmt.Println(err)
	}
	if err := json.Unmarshal(jsonBytes, &resources); err != nil {
		fmt.Println(err)
	}

	var (
		resourceList []string
		resourceType map[string]string
	)
	switch code {
	case "telegram":
		resourceType = resources.Telegram
	case "links":
		resourceType = resources.Links
	case "interest":
		resourceType = resources.InterestGroups
	}

	for k, v := range resourceType {
		resourceList = append(resourceList, fmt.Sprintf("%v : %v", k, v))
	}
	sort.Strings(resourceList)

	return "*" + code + "*\n" + strings.Join(resourceList, "\n")
}

//Structs for weather forecast function
type WeatherForecast struct {
	AM []AreaMetadata `json:"area_metadata"`
	FD []ForecastData `json:"items"`
}

type AreaMetadata struct {
	Name string            `json:"name"`
	Loc  tgbotapi.Location `json:"label_location"`
}

type ForecastData struct {
	FMD []ForecastMetadata `json:"forecasts"`
}

type ForecastMetadata struct {
	Name     string `json:"area"`
	Forecast string `json:"forecast"`
}

//Weather checks the weather based on given location
func (cb *Cinnabot) Weather(msg *message) {
	//Check if weather was sent with location, if not reply with markup
	if len(msg.Args) == 0 || !cb.CheckArgCmdPair("/weather", msg.Args) {
		opt1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Cinnamon"))
		opt2B := tgbotapi.NewKeyboardButton("Here")
		opt2B.RequestLocation = true
		opt2 := tgbotapi.NewKeyboardButtonRow(opt2B)

		options := tgbotapi.NewReplyKeyboard(opt1, opt2)

		replyMsg := tgbotapi.NewMessage(int64(msg.Message.From.ID), "🤖: Where are you?\n\n")
		replyMsg.ReplyMarkup = options
		cb.SendMessage(replyMsg)
		return
	}

	//Default loc: Cinnamon
	loc := &tgbotapi.Location{Latitude: 1.306671, Longitude: 103.773556}

	if msg.Location != nil {
		loc = msg.Location
	}

	//Send request to api.data.gov.sg for weather data
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://api.data.gov.sg/v1/environment/2-hour-weather-forecast", nil)
	req.Header.Set("api-key", "d1Y8YtThOpkE5QUfQZmvuA3ktrHa1uWP")

	resp, _ := client.Do(req)
	responseData, _ := ioutil.ReadAll(resp.Body)

	wf := WeatherForecast{}
	if err := json.Unmarshal(responseData, &wf); err != nil {
		log.Fatal(err)
		return
	}

	lowestDistance := distanceBetween(wf.AM[0].Loc, *loc)
	nameMinLoc := wf.AM[0].Name
	for i := 1; i < len(wf.AM); i++ {
		currDistance := distanceBetween(wf.AM[i].Loc, *loc)
		if currDistance < lowestDistance {
			lowestDistance = currDistance
			nameMinLoc = wf.AM[i].Name
		}
	}
	log.Print("The closest location is " + nameMinLoc)

	var forecast string
	for i, _ := range wf.FD[0].FMD {
		if wf.FD[0].FMD[i].Name == nameMinLoc {
			forecast = wf.FD[0].FMD[i].Forecast
			break
		}
	}

	//Parsing forecast
	words := strings.Fields(forecast)
	if len(words) > 1 {
		forecast = strings.ToLower(strings.Join(words[:len(words)-1], " "))
	} else {
		forecast = strings.ToLower(forecast)
	}

	responseString := "🤖: The 2h forecast is " + forecast + " for " + nameMinLoc
	returnMsg := tgbotapi.NewMessage(msg.Chat.ID, responseString)
	returnMsg.ParseMode = "Markdown"
	returnMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	cb.SendMessage(returnMsg)

}

//Helper funcs for weather
func distanceBetween(Loc1 tgbotapi.Location, Loc2 tgbotapi.Location) float64 {
	x := math.Pow((float64(Loc1.Latitude - Loc2.Latitude)), 2)
	y := math.Pow((float64(Loc1.Longitude - Loc2.Longitude)), 2)
	return x + y
}

// function to count number of users and messages
func (cb *Cinnabot) GetStats(msg *message) {

	db := model.InitializeDB()

	if cb.CheckArgCmdPair("/stats", msg.Args) {
		key := msg.Args[0]
		countUsers, countMessages := db.CountUsersAndMessages(key)
		mostUsedCommand := db.GetMostUsedCommand(key)

		extraString := ""
		if key != "forever" {
			extraString = " for the " + key
		}

		cb.SendTextMessage(int(msg.From.ID), "🤖: Here are some stats"+
			extraString+"!\n\n"+
			"Number of users registered on bot: "+strconv.Itoa(countUsers)+"\n"+
			"Numbery of messages typed: "+strconv.Itoa(countMessages)+"\n"+
			"Most used command: "+mostUsedCommand)
		return
	}

	opt1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Week"))
	opt2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Month"))
	opt3 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Year"))
	opt4 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Forever"))

	options := tgbotapi.NewReplyKeyboard(opt1, opt2, opt3, opt4)

	replyMsg := tgbotapi.NewMessage(int64(msg.Message.From.ID),
		"🤖: Please select the time period.")
	replyMsg.ReplyMarkup = options
	cb.SendMessage(replyMsg)

	return
}
