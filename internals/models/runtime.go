package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid format for runtime")

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	var s string
	if err := json.Unmarshal(jsonValue, &s); err != nil {
		return err
	}

	// Get rid of the mins part
	parts := strings.Split(s, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Parse the runtime int
	runtime, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(runtime)

	return nil
}
