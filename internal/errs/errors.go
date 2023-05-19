package errs

import (
	"encoding/json"
)

type DetailError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value,omitempty"`
	Msg      string `json:"msg"`
}

type DetailErrors struct {
	Errors []DetailError `json:"errors"`
	Status int           `json:"-"`
}

type MsgError struct {
	Msg    string `json:"message"`
	Status int    `json:"-"`
}

func (e DetailErrors) Error() string {
	result, err := json.Marshal(e)
	if err != nil {
		return "encode to json err: struct DetailErrors"
	}
	return string(result)
}

func (e MsgError) Error() string {
	result, err := json.Marshal(e)
	if err != nil {
		return "encode to json err: struct MsgError"
	}
	return string(result)
}
