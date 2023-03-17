package bot

import (
	"fmt"
	"github.com/elek/spiridon/db"
	"github.com/rs/zerolog/log"
	"strings"
)

type Notification struct {
	telegram      *Telegram
	subscriptions *db.Subscriptions
}

func NewNotification(telegram *Telegram, subscriptions *db.Subscriptions) *Notification {
	return &Notification{telegram: telegram, subscriptions: subscriptions}
}

func (n *Notification) Notify(node db.Node, failed []string, recovered []string) {
	subs, err := n.subscriptions.GetTargets(db.NodeSubcription, node.ID.String())
	if err != nil {
		log.Error().Err(err)
	}

	wSubs, err := n.subscriptions.GetTargets(db.WalletSubscription, node.OperatorWallet)
	if err != nil {
		log.Error().Err(err)
	}
	subs = append(subs, wSubs...)

	msg := fmt.Sprintf("Problem with node %s\n\n", node.ID)
	if len(recovered) > 0 {
		msg = "the following checks are recovered: " + strings.Join(recovered, ",") + "on node https://spiridon.anzix.net/node/" + node.ID.String()
		if len(failed) > 0 {
			msg += " Unfortunately we also have failures: " + strings.Join(failed, ",") + "."
		}
	} else if len(failed) > 0 {
		msg = "the following checks are failed: " + strings.Join(failed, ",") + " on node https://spiridon.anzix.net/node/" + node.ID.String()
	}

	for _, s := range subs {
		if s.DestinationType == db.TelegramSubscription {
			err := n.telegram.Send(s.Destination, msg)
			if err != nil {
				log.Error().Err(err).Msg("Couldn't send out Telegram notification")
			}
		}
	}
}
