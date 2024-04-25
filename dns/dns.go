// based on https://github.com/ahmetb/coredns-grpc-backend-sample/blob/master/server.go
// under Apache License:https://github.com/ahmetb/coredns-grpc-backend-sample/blob/master/LICENSE

package dns

import (
	"context"
	"fmt"
	"github.com/elek/spiridon/db"
	mdns "github.com/miekg/dns"
	"net"
	"storj.io/common/storj"
	"strings"
)

type Service struct {
	Db *db.Persistence
	UnimplementedDnsServiceServer
}

func NewService(db *db.Persistence) *Service {
	return &Service{
		Db: db,
	}
}

func (d *Service) Query(ctx context.Context, in *DnsPacket) (*DnsPacket, error) {
	m := new(mdns.Msg)
	if err := m.Unpack(in.Msg); err != nil {
		return nil, fmt.Errorf("failed to unpack msg: %v", err)
	}
	r := new(mdns.Msg)
	r.SetReply(m)
	r.Authoritative = true

	for _, q := range r.Question {
		hdr := mdns.RR_Header{Name: q.Name, Rrtype: q.Qtype, Class: q.Qclass}
		switch q.Qtype {
		case mdns.TypeA:
			id, err := storj.NodeIDFromString(strings.Split(q.Name, ".")[0])
			if err != nil {
				continue
			}
			get, err := d.Db.Get(db.NodeID{
				NodeID: id,
			})
			if err != nil {
				continue
			}
			ip := strings.Split(get.Address, ":")[0]
			ipAddr, err := net.ResolveIPAddr("ip", ip)
			if err != nil {
				continue
			}
			r.Answer = append(r.Answer, &mdns.A{Hdr: hdr, A: ipAddr.IP})
		default:
			return nil, fmt.Errorf("only A/AAAA supported, got qtype=%d", q.Qtype)
		}
	}

	if len(r.Answer) == 0 {
		r.Rcode = mdns.RcodeNameError
	}

	out, err := r.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack msg: %v", err)
	}
	return &DnsPacket{Msg: out}, nil
}
