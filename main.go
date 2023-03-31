package main

import (
	satellite "github.com/elek/spiridon/server"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	c := cobra.Command{
		Use: "spiridon",
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		//tp, err := initTracer()
		//if err != nil {
		//	return errors.WithStack(err)
		//}
		//defer func() {
		//	if err := tp.Shutdown(context.Background()); err != nil {
		//		log.Printf("Error shutting down tracer provider: %v", err)
		//	}
		//}()

		promExporter, err := otelprom.New()
		if err != nil {
			return err
		}

		metric.NewMeterProvider(
			metric.WithReader(promExporter),
		)

		//otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		//sdktrace.NewTracerProvider()

		var k = koanf.New(".")
		err = k.Load(confmap.Provider(map[string]interface{}{
			"web_port":  1234,
			"drpc_port": 7777,
			"db":        "postgres://postgres@localhost:5432/spiridon",
		}, "."), nil)
		if err != nil {
			return err
		}
		cfg := "spiridon.yml"
		if len(args) > 0 {
			cfg = args[0]
		}
		err = k.Load(file.Provider(cfg), yaml.Parser())
		if err != nil {
			return err
		}
		config := satellite.Config{}
		err = k.Unmarshal("", &config)
		if err != nil {
			return errors.WithStack(err)
		}
		return satellite.Run(config)
	}
	err := c.Execute()
	if err != nil {
		log.Error().Err(err).Msg("FAILED")
	}
}
