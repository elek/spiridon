package db

import (
	"github.com/pkg/errors"
	"time"
)

type Stat struct {
	Name   string
	NodeID NodeID
	Values []Measurement
}

type Measurement struct {
	Received time.Time
	Value    int
}

func (n *Persistence) LatestStat(nodeID NodeID) (time.Time, error) {
	var res time.Time
	rows, err := n.db.Raw("select date_trunc('month',received),key,field,max(value) from telemetries where node_id = ?   AND field in ('recent','sum') AND received > current_timestamp - interval '32 days' group by date_trunc('month',received),key,field order by key desc", nodeID).Rows()
	if err != nil {
		return res, errors.WithStack(err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&res)
	}
	if err != nil {
		return res, errors.WithStack(err)
	}
	return res, nil

}

func (n *Persistence) StateUpDown(nodeID NodeID, key string) (Stat, error) {
	s := Stat{
		Values: make([]Measurement, 0),
	}
	rows, err := n.db.Raw("select p.period,coalesce(max(value),0) FROM (select date_trunc('hour',generate_series(CURRENT_TIMESTAMP - INTERVAL '25 hours', CURRENT_TIMESTAMP,'1 hour'::interval)) as period) p "+
		"LEFT JOIN telemetries on period = date_trunc('hour',telemetries.received) AND  key =? and field ='count' and node_id = ? group by p.period order by p.period asc", key, nodeID).Rows()
	if err != nil {
		return s, errors.WithStack(err)
	}
	defer rows.Close()

	var received time.Time
	var val int

	for rows.Next() {
		err := rows.Scan(&received, &val)
		if err != nil {
			return s, errors.WithStack(err)
		}
		s.Values = append(s.Values, Measurement{
			Received: received,
			Value:    val,
		})
	}
	return s, nil
}
