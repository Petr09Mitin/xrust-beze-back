package custom_errors

import "encoding/json"

type CustomError struct {
	Msg string `json:"msg"`
}

func NewCustomError(msg string) *CustomError {
	return &CustomError{
		Msg: msg,
	}
}

func (e *CustomError) Error() string {
	return e.Msg
}

func (e *CustomError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"error": e.Error()})
}
