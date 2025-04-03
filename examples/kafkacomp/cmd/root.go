package cmd

import (
	"context"
	"encoding/json"
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
	SendMessage(ctx context.Context, topic string, key, value []byte) error
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
		kafkaComponent.SendMessage(context.Background(), "test_topic", []byte("key"), []byte("Hello, Kafka!"))
		slog.Info("Message sent successfully")

		// send a message with a struct
		type Message struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Email string `json:"email"`
		}
		msg := Message{
			Name:  "John Doe",
			Age:   30,
			Email: "bot@abc.com",
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			slog.Error("failed to marshal message", "error", err)
			panic(err)
		}
		err = kafkaComponent.SendMessage(context.Background(), "test_topic", []byte("key_object"), msgBytes)
		if err != nil {
			slog.Error("failed to send message", "error", err)
			panic(err)
		}
		slog.Info("Message sent successfully", "message", msg)

	},
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
