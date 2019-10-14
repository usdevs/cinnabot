package cinnabot

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// Wrapper struct for a callback query
type callback struct {
	ChatID int64
	MsgID  int
	Cmd    string
	Args   []string
	*tgbotapi.CallbackQuery
}

// GetArgStrings prints out the arguments for the callback in one string.
func (qry callback) GetArgString() string {
	return strings.Join(qry.Args, " ")
}

// CallbackFunc is a handler for a callback function
type CallbackFunc func(*callback)

// AddHandler binds a handler function to a callback cmd string in Cinnabot's HandlerMap
func (cb *Cinnabot) AddHandler(command string, resp CallbackFunc) error {
	if !strings.HasPrefix(command, "//") {
		return fmt.Errorf("not a valid callback string - it should be of the format //cmd [args]")
	}
	cb.hmap[command] = resp
	return nil
}

// Handle routes Telegram callback queries to the appropriate handlers.
func (cb *Cinnabot) Handle(qry tgbotapi.CallbackQuery) {
	// Parse callback
	parsed := cb.parseCallback(&qry)

	// Check that callback command exists
	execHandler, ok := cb.hmap[parsed.Cmd]
	if !ok {
		cb.log.Printf("[%s][id: %d] callback %s not registered!", time.Now().Format(time.RFC3339), parsed.ChatID, parsed.Cmd)
		cb.SendTextMessage(int(parsed.ChatID), "An error has occured! please notify the developers")
		return
	}

	// Log and execute callback
	cb.log.Printf("[%s][id: %d] callback: %s, args: %s", time.Now().Format(time.RFC3339), parsed.ChatID, parsed.Cmd, parsed.GetArgString())
	log.Print(parsed.ChatID)
	cb.GoSafely(func() { execHandler(parsed) })
}

// Helper to parse callbacks from inline keyboards
func (cb *Cinnabot) parseCallback(qry *tgbotapi.CallbackQuery) *callback {
	// Should add some error checking
	chatID := qry.Message.Chat.ID
	MsgID := qry.Message.MessageID
	qryTokens := strings.Fields(qry.Data)
	cmd, args := strings.ToLower(qryTokens[0]), qryTokens[1:]
	return &callback{ChatID: chatID, MsgID: MsgID, Cmd: cmd, Args: args, CallbackQuery: qry}
}

// InitCinnabot initializes an instance of Cinnabot.
// func InitCinnabotPatch(configJSON []byte, lg *log.Logger) *Cinnabot {
// 	if lg == nil {
// 		lg = log.New(os.Stdout, "[Cinnabot] ", 0)
// 	}
//
// 	cb := &Cinnabot{
// 		hmap:     make(map[string]CallbackFunc),
// 		log:      lg,
// 		cache:    cache.New(1*time.Minute, 2*time.Minute),
// 		Cinnabot: InitCinnabot(configJSON, lg),
// 	}
//
// 	return cb
// }
