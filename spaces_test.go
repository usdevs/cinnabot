package cinnabot

import (
	"testing"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

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

	expectedMsgStr1 := "Displaying bookings on Mon 19 Nov 2018:\n\n" +
		"=======================\nChatterbox\n=======================\nIntersection of Tradition & Technology: Japan Info Session\n07:00PM to 08:00PM, Mon 19 Nov 2018\n\nSem 2 Elections Open Discussion\n08:00PM to 10:00PM, Mon 19 Nov 2018\n\n" +
		"=======================\nUSP Master's Common\n=======================\n\"Owning Shakespeare: Scholars vs Actors.\" by Professor Michael Dobson\n06:30PM to 09:00PM, Mon 19 Nov 2018\n\n"

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
	expectedMsgStr := "Displaying bookings on Fri 23 Nov 2018:\n\n[No bookings recorded]"
	expectedMsg := makeExpectedMessage(999, expectedMsgStr)
	mb.On("Send", expectedMsg).Return(nil)
	cb.Spaces(&mockMsg)
}
