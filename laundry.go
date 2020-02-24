package cinnabot

import (
	"encoding/json"
	"fmt"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
	"sort"
	"strings"
	"time"

	fs "github.com/usdevs/cinnabot/firestore"
	"github.com/usdevs/cinnabot/utils"
)

// Backend

const projectId = "usc-laundry-test"

// Firestore document structs
type machineData struct {
	PinId              fs.String `json:"pinId"`
	Name               fs.String `json:"name"`
	Ezlink             fs.Bool   `json:"ezlink"`
	Washer             fs.Bool   `json:"washer"`
	On                 fs.Bool   `json:"on"`
	TimeChanged        fs.Time   `json:"timeChanged"`
	TimeChangedCertain fs.Bool   `json:"timeChangedCertain"`
	Cinnabot           fs.Bool   `json:"cinnabot"`
	Pi                 fs.Int    `json:"piNo"`
}

type piData struct {
	PiNo     fs.Int  `json:"piNo"`
	Level    fs.Int  `json:"level"`
	LastSeen fs.Time `json:"lastSeen"`
}

func firestoreToMachine(machineD machineData, pi piData) machine {
	return machine{
		Name:               machineD.Name.Value(),
		Ezlink:             machineD.Ezlink.Value(),
		Washer:             machineD.Washer.Value(),
		On:                 machineD.On.Value(),
		TimeChanged:        machineD.TimeChanged.Value(),
		TimeChangedCertain: machineD.TimeChangedCertain.Value(),
		Level:              pi.Level.Value(),
		LastSeen:           pi.LastSeen.Value(),
	}
}

func machineQuery() fs.Query {
	return fs.Query{
		From: []fs.CollectionSelector{{
			CollectionId:   "laundry_status",
			AllDescendants: false,
		}},
		Where: make(fs.Filters, 0),
	}
}

func getPiData() (map[int]piData, error) {
	query := fs.Query{
		From: []fs.CollectionSelector{{
			CollectionId:   "pi_status",
			AllDescendants: false,
		}},
		Where: make(fs.Filters, 0),
	}
	parse := func(data json.RawMessage) (interface{}, error) {
		obj := piData{}
		err := json.Unmarshal(data, &obj)
		return obj, err
	}
	piDocs, err := fs.RunQueryAndParse(projectId, query, parse, false)

	piMap := make(map[int]piData, len(piDocs))
	if err != nil {
		log.Print("Error getting pi data from firestore:", err)
		return piMap, err
	}
	for _, doc := range piDocs {
		data := doc.Data.(piData)
		piMap[data.PiNo.Value()] = data
	}
	return piMap, nil
}

func getMachineData(query fs.Query) ([]machineData, error) {
	parse := func(data json.RawMessage) (interface{}, error) {
		obj := machineData{}
		err := json.Unmarshal(data, &obj)
		return obj, err
	}
	machineDocs, err := fs.RunQueryAndParse(projectId, query, parse, false)

	machines := make([]machineData, 0, len(machineDocs))
	if err != nil {
		log.Print("Error getting laundry machine data from firestore:", err)
		return machines, err
	}
	for _, doc := range machineDocs {
		machines = append(machines, doc.Data.(machineData))
	}
	return machines, nil
}

func toMachines(mData []machineData, piMap map[int]piData) []machine {
	machines := make([]machine, 0, 18)
	for _, md := range mData {
		if pi, ok := piMap[md.Pi.Value()]; ok {
			machines = append(machines, firestoreToMachine(md, pi))
		} else {
			log.Printf("toMachines function in laundry.go: pi %d, which is connected to machine %s, cannot be found in firestore", md.Pi, md.Name)
			// log.Printf("Machine %s is connected to pi %d which cannot be found in firestore", md.Name.Value(), md.Pi.Value())
		}
	}
	return machines
}

// splitByLevel returns the machines in level 9/level 17 respectively
func splitByLevel(machines []machine) ([]machine, []machine) {
	l9 := make([]machine, 0)
	l17 := make([]machine, 0)
	for _, m := range machines {
		if m.Level == 9 {
			l9 = append(l9, m)
		} else if m.Level == 17 {
			l17 = append(l17, m)
		} else {
			log.Printf("Washing machine with invalid level (level %d, name %s)", m.Level, m.Name)
		}
	}
	return l9, l17
}

func splitByType(machines []machine) ([]washer, []dryer) {
	washers := make([]washer, 0)
	dryers := make([]dryer, 0)
	for _, m := range machines {
		if m.Washer {
			washers = append(washers, washer(m))
		} else {
			dryers = append(dryers, dryer(m))
		}
	}
	return washers, dryers
}

// used by UI

func getAllMachines() ([]level, error) {
	levels := make([]level, 0, 2)
	piData, piErr := getPiData()
	if piErr != nil {
		return levels, piErr
	}
	mData, mErr := getMachineData(machineQuery())
	if mErr != nil {
		return levels, mErr
	}
	machines := toMachines(mData, piData)
	l9, l17 := splitByLevel(machines)
	for i, m := range [2][]machine{l9, l17} {
		var lastSeen time.Time
		if len(m) > 0 { // if there are no machines on a level it won't be printed
			lastSeen = m[0].LastSeen
		}
		washers, dryers := splitByType(m)
		var l int
		if i == 0 {
			l = 9
		} else {
			l = 17
		}
		levels = append(levels, level{l, washers, dryers, lastSeen})
	}
	return levels, nil
}

// func getMachineType(isWasher bool) []level {
// 	// not the most efficient solution but there are only 18 washers + dryers
// 	levels := getAllMachines()
// 	for _, l := range levels {
// 		if isWasher {
// 			l.dryers = make([]dryer, 0)
// 		} else {
// 			l.washers = make([]washer, 0)
// 		}
// 	}
// 	return levels
// }

// func getLevel(lvl int) []level {
// 	levels := make([]level, 0, 1)
// 	if lvl != 9 && lvl != 17 {
// 		log.Printf("Invalid level used in getLevel function in laundry.go: there's no laundry room on level %d", lvl)
// 		return levels
// 	}

// 	piData, piErr := getPiData()
// 	if piErr != nil {
// 		return levels
// 	}
// 	mData, mErr := getMachineData(machineQuery())
// 	if mErr != nil {
// 		return levels
// 	}
// 	machines := toMachines(mData, piData)
// 	l9, l17 := splitByLevel(machines)
// 	if lvl == 9 {
// 		machines = l9
// 	} else if lvl == 17 {
// 		machines = l17
// 	}
// 	washers, dryers := splitByType(machines)
// 	levels = append(levels, level{lvl, washers, dryers})
// 	return levels
// }

// UI

const (
	washerCycle      = time.Duration(time.Minute * 30)
	dryerSingleCycle = time.Duration(time.Minute * 45)
	dryerDoubleCycle = dryerSingleCycle * 2

	notWorkingStr = " (sensor may not be working)"
	notCertainStr = " (or less)"
)

func formatTimeLeft(timeLeft time.Duration) string {
	if timeLeft < 0 {
		timeLeft = 0
	}
	return fmt.Sprintf("%d mins left", int(timeLeft.Minutes()))
}

func seenRecently(lastSeen time.Time) bool {
	return time.Since(lastSeen) < time.Duration(time.Minute*10)
}

type machine struct {
	Name               string
	Level              int
	Ezlink             bool
	Washer             bool
	On                 bool
	TimeChanged        time.Time
	TimeChangedCertain bool
	LastSeen           time.Time
}

func (m machine) timeLeft(cycleLength time.Duration) time.Duration {
	return cycleLength - time.Since(m.TimeChanged)
}

func (m machine) description() string {
	var ezlinkStr string
	if m.Ezlink {
		ezlinkStr = "ezlink"
	} else {
		ezlinkStr = "coin"
	}
	return fmt.Sprintf("*%s (%s)*", m.Name, ezlinkStr)
}

type washer machine

func (w washer) timeLeft() time.Duration {
	return machine(w).timeLeft(washerCycle)
}

func (w washer) string() string {
	str := machine(w).description() + ": "
	if w.On {
		// not free
		timeLeft := w.timeLeft()
		str += formatTimeLeft(timeLeft)
		if timeLeft.Minutes() < -5 {
			str += notWorkingStr
		} else if !w.TimeChangedCertain {
			str += notCertainStr
		}
	} else {
		str += "free"
	}
	return str
	// treats the last seen time of each sensor individually
	// str := machine(w).description() + ": "
	// sensorWorking := seenRecently(w.LastSeen)
	// if w.On {
	// 	// not free
	// 	timeLeft := w.timeLeft()
	// 	str += formatTimeLeft(timeLeft)
	// 	if timeLeft < 0 && !sensorWorking || timeLeft.Minutes() < -5 {
	// 		str += notWorkingStr
	// 	} else if !w.TimeChangedCertain {
	// 		str += notCertainStr
	// 	}
	// } else {
	// 	str += "free"
	// 	if !sensorWorking {
	// 		str += notWorkingStr
	// 	}
	// }
	// return str
}

type dryer machine

func (d dryer) timeLeft() (time.Duration, time.Duration) {
	m := machine(d)
	return m.timeLeft(dryerSingleCycle), m.timeLeft(dryerDoubleCycle)
}

func (d dryer) string() string {
	str := machine(d).description() + ": "
	// sensorWorking := seenRecently(d.LastSeen)
	if d.On {
		// not free
		single, double := d.timeLeft()
		if single.Minutes() >= -2 { // 2 mins buffer
			str += formatTimeLeft(single) + " (single tap) / "
		}
		str += formatTimeLeft(double) + " (double tap)"
		if double.Minutes() < -5 { // single < 0 && !sensorWorking || double.Minutes() < -5 {
			str += notWorkingStr
		} else if !d.TimeChangedCertain {
			str += notCertainStr
		}
	} else {
		// free
		str += "free"
		// 	if !sensorWorking {
		// 		str += notWorkingStr
		// 	}
	}
	return str
}

// sorting

func genericLess(first, second machine) bool {
	if first.Name == second.Name {
		return first.Ezlink
	} else {
		return first.Name < second.Name
	}
}

func sortWashers(washers []washer) {
	less := func(i, j int) bool {
		return genericLess(machine(washers[i]), machine(washers[j]))
	}
	sort.Slice(washers, less)
}

func sortDryers(dryers []dryer) {
	less := func(i, j int) bool {
		return genericLess(machine(dryers[i]), machine(dryers[j]))
	}
	sort.Slice(dryers, less)
}

type level struct {
	lvl     int
	washers []washer
	dryers  []dryer
	// assumption: 1 RPi per floor
	piLastSeen time.Time
}

func (l level) Len() int {
	return len(l.washers) + len(l.dryers)
}

func (l level) string() string {
	sortWashers(l.washers)
	sortDryers(l.dryers)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*LEVEL %d*\n", l.lvl))

	if !seenRecently(l.piLastSeen) {
		txt := "\u2757 Probably not accurate, the sensors were last seen at " + l.piLastSeen.Format(time.RFC822) + ".\n"
		sb.WriteString(txt)
	}

	if len(l.washers) > 0 {
		sb.WriteString("====== *Washers* ======")
		for _, m := range l.washers {
			sb.WriteString("\n")
			sb.WriteString(m.string())
		}
		sb.WriteString("\n")
	}
	if len(l.dryers) > 0 {
		sb.WriteString("======= *Dryers* =======")
		for _, m := range l.dryers {
			sb.WriteString("\n")
			sb.WriteString(m.string())
		}
	}
	return sb.String()
}

func laundryMsg() string {
	lastUpdated := "Last updated: " + time.Now().In(utils.SgLocation()).Format(time.RFC822)
	levels, err := getAllMachines()

	if err != nil {
		return "Could not fetch laundry availability.\n" + lastUpdated
	}

	toSend := ""
	for _, l := range levels {
		if l.Len() > 0 {
			toSend += l.string()
			toSend += "\n\n"
		}
	}

	return toSend + lastUpdated
}

func makeLaundryRefreshButton() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Refresh", "//laundry_refresh")),
	)
}

func (cb *Cinnabot) LaundryRefresh(qry *callback) {
	text := laundryMsg()
	refreshButton := makeLaundryRefreshButton()
	msg := tgbotapi.NewEditMessageText(qry.ChatID, qry.MsgID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = &refreshButton
	cb.SendMessage(msg)
}

// Laundry checks the washer and dryer availability.
func (cb *Cinnabot) Laundry(msg *message) {
	text := laundryMsg()
	refreshButton := makeLaundryRefreshButton()
	newMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	newMsg.ParseMode = tgbotapi.ModeMarkdown
	newMsg.ReplyMarkup = &refreshButton
	cb.SendMessage(newMsg)
}
