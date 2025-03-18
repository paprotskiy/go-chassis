package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	. "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	verifier *Validate
)

func init() {
	verifier = New()
}

type empty interface {
	Empty() bool
}

func GenericParse[T any](r *http.Request, disallowUnknownFields bool, postHandlers ...func(in *T)) (t *T, er error) {
	defer func() {
		if r := recover(); r != nil {

			er = errors.Errorf("panic has occurred: %s", r)
		}
	}()

	var tmp any = *new(T)
	if _, ok := tmp.(empty); ok {
		return new(T), nil
	}
	if _, ok := tmp.(struct{}); ok {
		return new(T), nil
	}

	data := new(T)
	decoder := json.NewDecoder(r.Body)
	if disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	err := decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode body :: %v", err)
	}

	for _, handler := range postHandlers {
		handler(data)
	}

	if err := validate(*data); err != nil {
		return nil, fmt.Errorf("failed to verify result :: %v", err.Error())
	}

	return data, nil
}

func validate(in any) error {
	kind := reflect.TypeOf(in).Kind()
	switch kind {
	case reflect.Slice, reflect.Array, reflect.Map:
		return verifier.Var(in, "required,dive")
	case reflect.Struct:
		return verifier.Struct(in)
	case reflect.Ptr:
		elem := reflect.ValueOf(in).Elem()
		return validate(elem)
	default:
		return fmt.Errorf("unexpected kind of type :: %v", kind.String())
	}
}
