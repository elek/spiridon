package ops

import (
	"context"
	"fmt"
	"github.com/spacemonkeygo/monkit/v3"
	"net"
	"net/http"
	"strings"
)

type Metrics struct {
	listener net.Listener
}

func NewMetrics() (*Metrics, error) {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:4444")
	if err != nil {
		return nil, err
	}
	return &Metrics{
		listener: tcpListener,
	}, nil
}

func (d *Metrics) Handle(w http.ResponseWriter, r *http.Request) {
	data := make(map[string][]string)
	var components []string
	monkit.Default.Stats(func(key monkit.SeriesKey, field string, val float64) {
		components = components[:0]

		measurement := sanitize(key.Measurement)
		for tag, tagVal := range key.Tags.All() {
			components = append(components,
				fmt.Sprintf("%s=%q", sanitize(tag), sanitize(tagVal)))
		}
		components = append(components,
			fmt.Sprintf("field=%q", sanitize(field)))

		data[measurement] = append(data[measurement],
			fmt.Sprintf("{%s} %g", strings.Join(components, ","), val))
	})

	for measurement, samples := range data {
		_, _ = fmt.Fprintln(w, "# TYPE", measurement, "gauge")
		for _, sample := range samples {
			_, _ = fmt.Fprintf(w, "%s%s\n", measurement, sample)
		}
	}
}

func (d *Metrics) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", http.HandlerFunc(d.Handle))
	return http.Serve(d.listener, mux)
}

// sanitize formats val to be suitable for prometheus.
func sanitize(val string) string {
	// https://prometheus.io/docs/concepts/data_model/
	// specifies all metric names must match [a-zA-Z_:][a-zA-Z0-9_:]*
	// Note: The colons are reserved for user defined recording rules.
	// They should not be used by exporters or direct instrumentation.
	if val == "" {
		return ""
	}
	if '0' <= val[0] && val[0] <= '9' {
		val = "_" + val
	}
	return strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z':
			return r
		case 'A' <= r && r <= 'Z':
			return r
		case '0' <= r && r <= '9':
			return r
		default:
			return '_'
		}
	}, val)
}
