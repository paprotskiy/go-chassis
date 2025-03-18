package msgbroker

import (
	"reflect"

	"github.com/pkg/errors"
)

func newSubscribingErr(err error) error {
	return &SubscribingErr{
		msg: errors.Wrap(err, "failed to subscribe to topic").Error(),
	}
}

type SubscribingErr struct {
	msg string
}

func (s *SubscribingErr) Error() string {
	return s.msg
}

func newMsgReadingErr(err error, topic string) error {
	return &MsgReadingErr{
		msg: errors.Wrapf(err, `failed to read message from topic "%s"`, topic).Error(),
	}
}

type MsgReadingErr struct {
	msg string
}

func (s *MsgReadingErr) Error() string {
	return s.msg
}

func newConversionErr[T any](err error) error {
	typeof := reflect.TypeOf(*new(T)).String()
	return &ConversionErr{
		msg: errors.Wrapf(err, "failed to convert data to %s", typeof).Error(),
	}
}

type ConversionErr struct {
	msg string
}

func (s *ConversionErr) Error() string {
	return s.msg
}

func newProcessingErr(err error) error {
	return &ProcessingErr{
		msg: errors.Wrap(err, "failed to process converted message").Error(),
	}
}

type ProcessingErr struct {
	msg string
}

func (s *ProcessingErr) Error() string {
	return s.msg
}

func newCommitMsgErr(err error) error {
	return &CommitMsgErr{
		msg: errors.Wrap(err, "failed to commit message").Error(),
	}
}

type CommitMsgErr struct {
	msg string
}

func (s *CommitMsgErr) Error() string {
	return s.msg
}
