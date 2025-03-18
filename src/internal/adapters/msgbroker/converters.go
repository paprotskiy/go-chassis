package msgbroker

import "github.com/confluentinc/confluent-kafka-go/kafka"

func converterStub(in *kafka.Message) (*any, error) {
	return nil, nil
}
