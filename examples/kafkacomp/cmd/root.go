package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"github.com/taimaifika/service-context/component/kafkac"
	"github.com/taimaifika/service-context/component/slogc"

	sctx "github.com/taimaifika/service-context"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("kafka-example"),
		sctx.WithComponent(slogc.NewSlogComponent()),
		sctx.WithComponent(kafkac.NewKafkaComponent("kafka")),
	)
}

type KafkaComponent interface {
	GetProducer() *sarama.SyncProducer
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start gin service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		if err := serviceCtx.Load(); err != nil {
			slog.Error("load service context error", "error", err)
			panic(err)
		}

		// Get the Kafka component
		kafkaComponent := serviceCtx.MustGet("kafka").(KafkaComponent)
		producer := kafkaComponent.GetProducer()
		if producer == nil {
			slog.Error("producer is nil")
			panic("producer is nil")
		}

		// Use the producer to send a message
		msg := &sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("Hello, Kafka!"),
		}
		partition, offset, err := (*producer).SendMessage(msg)
		if err != nil {
			slog.Error("failed to send message", "error", err)
			panic(err)
		}

		slog.Info("Message sent successfully", "partition", partition, "offset", offset)
	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
