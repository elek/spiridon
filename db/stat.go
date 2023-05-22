package db

import (
	"context"
	"github.com/pkg/errors"
	"github.com/spacemonkeygo/monkit/v3"
	"time"
)

var mon = monkit.Package()

type Stat struct {
	Name   string
	NodeID NodeID
	Values []Measurement
}

type Measurement struct {
	Received time.Time
	Value    int
}

type StatCollection struct {
	Records []StatRecord
}

func (c *StatCollection) Insert(record StatRecord) {
	for _, r := range c.Records {
		if r.Key == record.Key && r.Field == record.Field {
			return
		}
	}
	c.Records = append(c.Records, record)
}

func (c *StatCollection) Get(key string, field string) StatRecord {
	for _, r := range c.Records {
		if r.Key == key && r.Field == field {
			return StatRecord{
				Received: r.Received,
				Key:      r.Key,
				Value:    r.Value,
				Field:    r.Field,
				Actual:   true,
			}
		}
	}
	return StatRecord{
		Received: time.UnixMicro(0),
		Key:      key,
		Field:    field,
		Value:    0,
		Actual:   false,
	}
}

type NodeStat struct {
	Enabled         bool
	UsedSpace       StatRecord
	UploadedBytes   StatRecord
	DownloadedBytes StatRecord
}

type StatRecord struct {
	Received time.Time
	Key      string
	Field    string
	Value    float64
	Actual   bool
}

func (c *StatCollection) AsNodeStat() NodeStat {
	return NodeStat{
		Enabled:   len(c.Records) > 0,
		UsedSpace: c.Get("scope=storj.io/storj/storagenode/monitor", "recent"),
	}
}

func (n *Persistence) LatestStat(ctx context.Context, nodeID NodeID) (stat StatCollection, err error) {
	defer mon.Task()(&ctx)(&err)
	rows, err := n.db.Raw("select b.received,b.key,b.field,b.value from telemetries b JOIN (select key,field,max(received) as received from telemetries where field in ('sum','recent') group by key,field) r ON r.field=b.field AND r.key=b.key AND r.received = b.received WHERE node_id=?;", nodeID).Rows()
	if err != nil {
		return stat, errors.WithStack(err)
	}
	defer rows.Close()

	for rows.Next() {
		var rec StatRecord
		err = rows.Scan(&rec.Received, &rec.Key, &rec.Field, &rec.Value)
		if err != nil {
			return stat, errors.WithStack(err)
		}
		stat.Insert(rec)
	}

	return stat, nil

}

func (n *Persistence) GetStat(ctx context.Context, nodeID NodeID, key string) (stat Stat, err error) {
	defer mon.Task()(&ctx)(&err)
	stat = Stat{
		Values: make([]Measurement, 0),
	}
	rows, err := n.db.Raw("select p.period,coalesce(max(value),0) FROM (select date_trunc('hour',generate_series(CURRENT_TIMESTAMP - INTERVAL '25 hours', CURRENT_TIMESTAMP,'1 hour'::interval)) as period) p "+
		"LEFT JOIN telemetries on period = date_trunc('hour',telemetries.received) AND  key =? and field ='count' and node_id = ? group by p.period order by p.period asc", key, nodeID).Rows()
	if err != nil {
		return stat, errors.WithStack(err)
	}
	defer rows.Close()

	var received time.Time
	var val int

	for rows.Next() {
		err := rows.Scan(&received, &val)
		if err != nil {
			return stat, errors.WithStack(err)
		}
		stat.Values = append(stat.Values, Measurement{
			Received: received,
			Value:    val,
		})
	}
	return stat, err
}
