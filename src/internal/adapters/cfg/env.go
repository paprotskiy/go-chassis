package cfg

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type envReader struct {
	errMessages []string
}

func (e *envReader) parsingErr(errSeparator string) error {
	if len(e.errMessages) > 0 {
		concatenatedErrMsg := "\n" + strings.Join(e.errMessages, errSeparator)
		return errors.New(concatenatedErrMsg)
	}
	return nil
}

func (e *envReader) toString(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		e.errMessages = append(e.errMessages, fmt.Sprintf("env variable not found. Key : %v", key))
	}
	return value
}

func (e *envReader) toTimeDuration(key string) time.Duration {
	value := os.Getenv(key)
	if len(value) == 0 {
		e.errMessages = append(e.errMessages, fmt.Sprintf("env variable not found. Key : %v", key))
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		e.errMessages = append(e.errMessages, err.Error())
	}

	return duration
}

func (e *envReader) toUrl(key string) (parsedUrl url.URL) {
	value := os.Getenv(key)
	if len(value) == 0 {
		e.errMessages = append(e.errMessages, fmt.Sprintf("env variable not found. Key : %v", key))
		return
	}

	ref, err := url.Parse(value)
	if err != nil {
		e.errMessages = append(e.errMessages, fmt.Sprintf("failed to parse url by key %v :: %v", key, err.Error()))
		return
	}

	parsedUrl = *ref
	return
}

func (e *envReader) toInt(key string) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		e.errMessages = append(e.errMessages, fmt.Sprintf("env variable not found. Key : %v", key))
		return 0
	}

	res, err := strconv.Atoi(value)
	if len(value) == 0 {
		e.errMessages = append(e.errMessages, fmt.Sprintf("failed to parse int value : %v", err))
		return 0
	}

	return res
}

func (e *envReader) toBoolSupressErr(key string) bool {
	b, _ := strconv.ParseBool(os.Getenv(key))
	return b
}
