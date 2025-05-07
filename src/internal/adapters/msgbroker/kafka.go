package msgbroker

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

// TODO: add security layer

const (
	acksAll = "all" // ensure all in-sync
	acks1   = "1"   // acknowledgment from partition leader
	acks0   = "0"   // no acknowledgment from broker
)

func New(address string) (*kafkaAdapter, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": address,
		"acks":              acksAll,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new producer")
	}

	return &kafkaAdapter{
		address:  address,
		producer: producer,
	}, nil
}

type kafkaAdapter struct {
	address  string
	producer *kafka.Producer
}

func (k *kafkaAdapter) PushToBroker(data []byte) error {
	return nil // TODO: implement by wrapping k.produceMessage
}

func (k *kafkaAdapter) produceMessage(
	topic string,
	key, value []byte,
) error {
	deliveryChan := make(chan kafka.Event)
	err := k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(value),
		Key:            []byte(key),
	}, deliveryChan)
	if err != nil {
		return errors.Wrap(err, "failed to produce a message")
	}

	event := <-deliveryChan
	switch e := event.(type) {
	case *kafka.Message:
		if e.TopicPartition.Error != nil {
			return errors.Wrap(e.TopicPartition.Error, "failed to deliver message")
		}
	}

	return nil
}

func (k *kafkaAdapter) NewExampleConsumer(ctx context.Context, process func(*any) error, onErr func(error)) {
	newSuccessConsumer(
		ctx,
		k.address,
		"some.topic.v1.2.3",
		"group.id",
		converterStub,
		process,
		onErr,
	)
}

// warn: restart may be controlled through ctx for the first, but consumer requires improvements
// implements strategy of commiting after successful message processing
func newSuccessConsumer[T any](
	ctx context.Context,
	address string,
	topic string,
	groupId string,
	convert func(*kafka.Message) (*T, error),
	process func(*T) error,
	onErr func(error),
) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  address,
		"group.id":           groupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})

	if err != nil {
		onErr(errors.Wrap(err, "failed to create consumer"))
		return
	}

	subErr := c.Subscribe(topic, nil) // no rebalance
	if subErr != nil {
		onErr(newSubscribingErr(subErr))
		return
	}

	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
			msg, err := c.ReadMessage(-1)
			if err != nil {
				onErr(newMsgReadingErr(err, topic))
				continue
			}

			res, err := convert(msg)
			if err != nil {
				onErr(newConversionErr[T](err))
				continue
			}

			err = process(res)
			if err != nil {
				onErr(newProcessingErr(err))
				continue
			}

			_, err = c.CommitMessage(msg)
			if err != nil {
				onErr(newCommitMsgErr(err))
				continue
			}
		}
	}
}
