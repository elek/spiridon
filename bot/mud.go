package bot

import (
	"github.com/elek/spiridon/config"
	"github.com/elek/spiridon/mud"
)

func Module(ball *mud.Ball) {
	mud.Provide[*Notification](ball, NewNotification)
	mud.Provide[*Telegram](ball, func(cfg config.Config, robot *Robot) (*Telegram, error) {
		return NewTelegram(cfg.TelegramToken, robot)
	})
	mud.Provide[*Robot](ball, NewRobot)
}
