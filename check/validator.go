package check

import (
	"context"
	"fmt"
	"github.com/elek/spiridon/bot"
	"github.com/elek/spiridon/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"storj.io/common/identity"
	"strings"
	"time"
)

type Validator struct {
	checks       []Checker
	nodes        *db.Persistence
	notification *bot.Notification
}

func NewValidator(nodes *db.Persistence, notification *bot.Notification, identity *identity.FullIdentity) *Validator {
	return &Validator{
		nodes: nodes,
		checks: []Checker{
			&Port{identity: identity},
			&Quic{identity: identity},
			&Checkin{},
			NewHealthy(nodes),
			&Upload{identity: identity},
		},
		notification: notification,
	}
}

func (v *Validator) Loop(ctx context.Context) error {
	initialized := make(map[string]bool)
	for {
		nodes, err := v.nodes.ListNodesInternal()
		if err != nil {
			panic(err)
		}
		for _, node := range nodes {
			if running := initialized[node.ID.String()]; !running {
				go v.nodeLoop(node)
				initialized[node.ID.String()] = true
			}

		}
		time.Sleep(1 * time.Minute)
	}
}

func (v *Validator) nodeLoop(node db.Node) {
	for {
		v.checkNode(node)
		time.Sleep(1 * time.Minute)
	}
}

func (v *Validator) checkNode(node db.Node) {
	checked := map[string]db.CheckResult{}
	existing, err := v.nodes.GetStatus(node.ID)
	if err != nil {
		log.Error().Err(err)
	}
	var recovered []string
	var failed []string
	for _, check := range v.checks {
		result, found := existing[check.Name()]
		if !found || time.Since(result.LastChecked) > 5*time.Second {
			start := time.Now()
			err := check.Check(node)
			duration := time.Since(start)
			if err == nil || errors.Is(err, CheckinFailed) || errors.Is(err, CheckinWarning) {
				msg := ""
				if err != nil {
					msg = strings.TrimSuffix(strings.TrimSpace(err.Error()), ":")
				}
				checked[check.Name()] = db.CheckResult{
					Time:     time.Now(),
					Error:    msg,
					Duration: duration,
					Warning:  errors.Is(err, CheckinWarning),
				}
				if found {
					if result.Error == "" && msg != "" && !errors.Is(err, CheckinWarning) {
						failed = append(failed, check.Name())
					}
					if result.Error != "" && !result.Warning && msg == "" {
						recovered = append(recovered, check.Name())
					}
				}
			} else {
				log.Error().Err(err).Msg("Check has been failed")
			}
		}
	}
	if len(checked) > 0 {
		err := v.nodes.UpdateStatus(node.ID, checked)
		if err != nil {
			fmt.Println(err, node.ID)
		}
	}

	if len(failed) > 0 || len(recovered) > 0 {
		v.notification.Notify(node, failed, recovered)
	}
}
