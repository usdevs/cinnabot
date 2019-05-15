package cinnabot

import (
	"testing"

	"github.com/stretchr/testify/mock"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

//to run this test: go test spaces_test.go cinnabot.go spaces.go

type mockBot struct {
	mock.Mock
}

func (mb *mockBot) GetUpdatesChan(config tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	return nil, nil
}

func (mb *mockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := mb.Called(c)
	return tgbotapi.Message{}, args.Error(0)
}

// makeExpectedMessage returns MessageConfig with settings matching those sent by cinnabot
func makeExpectedMessage(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)

	//match settings of message sent by bot
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true, Selective: true}
	msg.ParseMode = "Markdown"
	return msg
}

func TestGetOnDate(t *testing.T) {
	mb := mockBot{}
	cb := Cinnabot{
		bot: &mb,
	}
	mockMsg1 := message{
		Args: []string{"19/11/18"},
		Message: &tgbotapi.Message{
			MessageID: 1,
			From: &tgbotapi.User{
				ID:        999,
				FirstName: "test_user_first_name",
			},
		},
	}

	expectedMsgStr1 := `Displaying all bookings on Mon 19 Nov 18:

=======================
Theme Room 1
=======================
*RA Internal Welfare Day:* 08PM, Sun 18 Nov 18 to 01AM, Mon 19 Nov 18

=======================
Chatterbox
=======================
*Intersection of Tradition & Technology: Japan Info Session:* 07PM to 08PM, Mon 19 Nov 18

*Sem 2 Elections Open Discussion:* 08PM to 10PM, Mon 19 Nov 18

=======================
USP Master's Common
=======================
*"Owning Shakespeare: Scholars vs Actors." by Professor Michael Dobson:* 06:30PM to 09PM, Mon 19 Nov 18

`

	expectedMsg1 := makeExpectedMessage(999, expectedMsgStr1)

	mb.On("Send", expectedMsg1).Return(nil)
	cb.Spaces(&mockMsg1)
}

func TestGetNoEvents(t *testing.T) {
	mb := mockBot{}
	cb := Cinnabot{
		bot: &mb,
	}
	mockMsg := message{
		Args: []string{"23/11/18"},
		Message: &tgbotapi.Message{
			MessageID: 1,
			From: &tgbotapi.User{
				ID:        999,
				FirstName: "test_user_first_name",
			},
		},
	}
	expectedMsgStr := "Displaying all bookings on Fri 23 Nov 18:\n\n[No bookings recorded]"
	expectedMsg := makeExpectedMessage(999, expectedMsgStr)
	mb.On("Send", expectedMsg).Return(nil)
	cb.Spaces(&mockMsg)
}
