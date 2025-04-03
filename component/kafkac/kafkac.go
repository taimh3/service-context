package kafkac

import (
	"context"
	"flag"
	"log/slog"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	AddrsStr string

	Addrs       []string
	maxRetries  int
	maxWaitTime time.Duration

	// Authentication
	SASLUser string
	SASLPass string
}

type kafkaComponent struct {
	id string

	*config

	producer *sarama.SyncProducer
}

func NewKafkaComponent(id string) *kafkaComponent {
	return &kafkaComponent{
		id:     id,
		config: new(config),
	}
}

func (k *kafkaComponent) ID() string {
	return k.id
}

func (k *kafkaComponent) InitFlags() {
	flag.StringVar(&k.AddrsStr, k.id+"-addrs", "localhost:9092", "kafka addresses. default: localhost:9092")
	flag.IntVar(&k.maxRetries, k.id+"-max-retries", 3, "kafka max retries. default: 3")
	flag.DurationVar(&k.maxWaitTime, k.id+"-max-wait-time", 10*time.Second, "kafka max wait time. default: 10s")

	flag.StringVar(&k.SASLUser, k.id+"-sasl-user", "", "kafka sasl user")
	flag.StringVar(&k.SASLPass, k.id+"-sasl-pass", "", "kafka sasl password")
}

func (k *kafkaComponent) Activate(ctx sctx.ServiceContext) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Retry.Max = k.maxRetries
	config.Producer.Timeout = k.maxWaitTime

	// set authentication
	if k.SASLUser != "" && k.SASLPass != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = k.SASLUser
		config.Net.SASL.Password = k.SASLPass
	}

	// Parse the addresses
	k.Addrs = strings.Split(k.AddrsStr, ",")

	// Create the producer
	slog.Info("Creating Kafka producer", "addresses", k.Addrs)
	producer, err := sarama.NewSyncProducer(k.Addrs, config)
	if err != nil {
		return err
	}

	k.producer = &producer
	slog.Info("Kafka producer created successfully", "addresses", k.Addrs)

	return nil
}

func (k *kafkaComponent) Stop() error {
	if k.producer != nil {
		slog.Info("Stopping Kafka producer")
		if err := (*k.producer).Close(); err != nil {
			slog.Error("Failed to close Kafka producer", "error", err)
		}
	}
	return nil
}

func (k *kafkaComponent) GetProducer() *sarama.SyncProducer {
	return k.producer
}

func (k *kafkaComponent) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	_, span := otel.Tracer("kafkaComponent").Start(ctx, "SendMessage")
	defer span.End()

	if k.producer == nil {
		return sarama.ErrNotConnected
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	slog.Info("Sending message to Kafka", "topic", topic, "key", string(key), "value", string(value))
	partition, offset, err := (*k.producer).SendMessage(msg)
	if err != nil {
		return err
	}

	slog.Info("Message sent successfully", "topic", topic, "partition", partition, "offset", offset)
	return nil
}

func (k *kafkaComponent) NewConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Create the consumer group
	slog.Info("Creating Kafka consumer group", "addresses", k.Addrs, "groupID", groupID)
	consumerGroup, err := sarama.NewConsumerGroup(k.Addrs, groupID, config)
	if err != nil {
		return nil, err
	}

	slog.Info("Kafka consumer group created successfully", "addresses", k.Addrs, "groupID", groupID)

	return consumerGroup, nil
}
