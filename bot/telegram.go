package bot

import (
	"fmt"
	"github.com/elek/spiridon/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
)

type Telegram struct {
	robot *Robot
	bot   *tgbotapi.BotAPI
}

func NewTelegram(token string, robot *Robot) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	bot.Debug = true
	return &Telegram{robot: robot, bot: bot}, nil
}

func (t *Telegram) Run() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			answer, err := t.robot.Handle(db.TelegramSubscription, fmt.Sprintf("%d", update.Message.Chat.ID), update.Message.Text)
			if err != nil {
				_, _ = t.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "ERROR:"+err.Error()))
				log.Err(err).Send()
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
			msg.ReplyToMessageID = update.Message.MessageID
			_, err = t.bot.Send(msg)
			if err != nil {
				log.Error().Err(err)
			}
		}
	}
	return nil
}

func (t *Telegram) Send(target string, msg string) error {
	channelID, err := strconv.ParseInt(target, 10, 64)
	if err != nil {
		return err
	}
	tmsg := tgbotapi.NewMessage(channelID, msg)
	_, err = t.bot.Send(tmsg)
	return err
}
