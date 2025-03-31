package kafkac

import (
	"flag"
	"log/slog"
	"strings"
	"time"

	"github.com/IBM/sarama"
	sctx "github.com/taimaifika/service-context"
)

type config struct {
	AddrsStr string

	Addrs       []string
	maxRetries  int
	maxWaitTime time.Duration
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
}

func (k *kafkaComponent) Activate(ctx sctx.ServiceContext) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Retry.Max = k.maxRetries
	config.Producer.Timeout = k.maxWaitTime

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
	return nil
}

func (k *kafkaComponent) GetProducer() *sarama.SyncProducer {
	return k.producer
}
