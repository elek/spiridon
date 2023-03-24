package bot

import (
	"fmt"
	"github.com/elek/spiridon/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"strings"
)

type Notification struct {
	telegram      *Telegram
	ntfy          *Ntfy
	subscriptions *db.Subscriptions
	db            *db.Nodes
}

func NewNotification(telegram *Telegram, subscriptions *db.Subscriptions, nodes *db.Nodes) *Notification {
	return &Notification{
		telegram:      telegram,
		subscriptions: subscriptions,
		db:            nodes,
		ntfy:          &Ntfy{}}
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

	wallet, err := n.db.GetWallet(common.HexToAddress(node.OperatorWallet))
	if err != nil {
		log.Error().Err(err)
	}
	if wallet.NtfyChannel != "" {
		subs = append(subs, db.Subscription{
			Destination:     wallet.NtfyChannel,
			DestinationType: db.NtfySubscription,
			Base:            wallet.Address,
			BaseType:        db.WalletSubscription,
		})
	}

	msg := fmt.Sprintf("Problem with node %s\n\n", node.ID)
	if len(recovered) > 0 {
		msg = "the following checks are recovered: " + strings.Join(recovered, ",") + " on node https://spiridon.anzix.net/node/" + node.ID.String()
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
		} else if s.DestinationType == db.NtfySubscription {
			err := n.ntfy.Send(s.Destination, msg)
			if err != nil {
				log.Error().Err(err).Msg("Couldn't send out Telegram notification")
			}
		}
	}
}
