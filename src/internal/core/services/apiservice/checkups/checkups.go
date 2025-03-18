package checkups

import "errors"

func ref[T any](in T) *T { return &in }

func SameType[CustomErr error](err error) bool {
	return errors.As(err, new(CustomErr))
}

type CheckupErr string

func (c *CheckupErr) Error() string {
	return string(*c)
}

func AlwaysErr() func() error {
	return func() error {
		return ref(CheckupErr("I am CheckupErr example"))
	}
}
