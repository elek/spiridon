package dns

import (
	"context"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestGrpcIntegration(t *testing.T) {
	serverAddr := "127.0.0.1:8053"
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewDnsServiceClient(conn)
	var msg dns.Msg
	msg.SetQuestion("example.org.", dns.TypeA)
	raw, err := msg.Pack()
	require.NoError(t, err)
	_, err = client.Query(context.Background(), &DnsPacket{
		Msg: raw,
	})
	require.NoError(t, err)
	defer conn.Close()
}
