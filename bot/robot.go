package bot

import (
	"fmt"
	"github.com/elek/spiridon/db"
	"storj.io/common/storj"
	"strings"
)

type Robot struct {
	nodes         *db.Nodes
	subscriptions *db.Subscriptions
}

func NewRobot(nodes *db.Nodes, subscriptions *db.Subscriptions) *Robot {
	return &Robot{nodes: nodes, subscriptions: subscriptions}
}

func (r *Robot) Handle(targetType int, target string, message string) (string, error) {
	message = strings.TrimSpace(message)
	message = strings.TrimPrefix(message, "/")
	first := strings.Split(message, " ")[0]
	switch first {
	case "start":
		return "Hi, I am the bot for https://spiridon.anzix.net. You can subscribe to Storage node status notification with `/subscribe <nodeid>`", nil
	case "hi":
		return "same to you", nil
	case "how":
		if message == "how are you?" {
			return "Fine thanks, and you?", nil
		}
	case "subscribe":
		help := "Use the format `/subscribe <nodeid>` to specify the node what you are interested about."
		parts := strings.SplitN(message, " ", 3)

		switch len(parts) {
		case 2:
			nodeID, err := storj.NodeIDFromString(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Sprintf("%s is not a valid NodeID", parts[1]), nil
			}

			err = r.subscriptions.Subscribe(db.Subscription{
				Destination:     target,
				DestinationType: targetType,
				Base:            parts[1],
				BaseType:        db.NodeSubcription,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("successfully subscribed to %s", nodeID.String()), err
		default:
			return help, nil
		}

	case "unsubscribe":
		parts := strings.SplitN(message, " ", 3)
		switch len(parts) {
		case 2:
			nodeID, err := storj.NodeIDFromString(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Sprintf("%s is not a valid NodeID", parts[1]), nil
			}

			err = r.subscriptions.Unsubscribe(db.Subscription{
				Destination:     target,
				DestinationType: targetType,
				Base:            parts[1],
				BaseType:        db.NodeSubcription,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("successfully unsubscribed from node %s", nodeID.String()), err
		}

	case "subscriptions", "list":

		subscriptions, err := r.subscriptions.ListSubscriptions(targetType, target)
		if err != nil {
			return "", err
		}
		if len(subscriptions) == 0 {
			return "You don't have any subscription yet. Use `/subscribe <nodeid>` to get notifications about node problems", nil
		}
		out := "Your current subscriptions:\n"
		for _, n := range subscriptions {
			switch n.BaseType {
			case db.NodeSubcription:
				out += fmt.Sprintf("%s (node)\n", n.Base)
			case db.WalletSubscription:
				out += fmt.Sprintf("%s (wallet)\n", n.Base)
			}

		}
		return out, nil
	case "size":
		nodes, err := r.nodes.ListNodes()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("I have %d number of nodes registered in my DB", len(nodes)), nil
	default:
		return "Sorry, I don't really understand you. Type help for the available commands", nil
	}
	return "", nil
}
