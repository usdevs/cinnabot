package main

import (
	"io/ioutil"
	"log"
	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	// "github.com/usdevs/cinnabot"
	"github.com/usdevs/cinnabot/model"
	"github.com/usdevs/cinnabot/callbacks"
)

func main() {
	configJSON, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("error reading config file! Boo: %s", err)
	}

	logger := log.New(os.Stdout, "[cinnabot] ", 0)

	cb := callbacks.InitCinnabotPatch(configJSON, logger)
	db := model.InitializeDB()

	//Junk functions
	cb.AddFunction("/echo", cb.Echo)
	cb.AddFunction("/hello", cb.SayHello)
	cb.AddFunction("/capitalize", cb.Capitalize)

	//Main functions
	cb.AddFunction("/start", cb.Start)
	cb.AddFunction("/about", cb.About)
	cb.AddFunction("/help", cb.Help)
	cb.AddFunction("/stats", cb.GetStats)

	cb.AddFunction("/resources", cb.Resources)
	cb.AddFunction("/publicbus", cb.BusTimings)
	cb.AddFunction("/nusbus", cb.NUSBus)
	cb.AddFunction("/weather", cb.Weather)

	cb.AddFunction("/spaces", cb.Spaces)

	cb.AddFunction("/feedback", cb.Feedback)
	cb.AddFunction("/dhsurvey", cb.DHSurvey)
	cb.AddFunction("/cinnabotfeedback", cb.CinnabotFeedback)
	cb.AddFunction("/uscfeedback", cb.USCFeedback)
	cb.AddFunction("/diningfeedback", cb.DiningFeedback)
	cb.AddFunction("/residentialfeedback", cb.ResidentialFeedback)
	cb.AddFunction("/dhsurveyfeedback", cb.DHSurveyFeedback)

	cb.AddFunction("/cancel", cb.Cancel)

	// Callback handlers
	// cb.addHandler("//refreshbustiming", cb.RefreshBusTiming)

	updates := cb.Listen(60)
	log.Println("Listening...")

	for update := range updates {
		if update.Message != nil {
			modelMsg, modelUsr := model.FromTelegramMessage(*update.Message)
			db.Add(&modelMsg)
			db.Add(&modelUsr)
			cb.Router(*update.Message)
		}
		if update.CallbackQuery != nil {
			cb.Handle(*update.CallbackQuery)
		}
	}

}
