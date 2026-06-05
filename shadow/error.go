package shadow

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("shadow not found")

// ErrorResponse represents an AWS IoT Core Device Shadow error response document, which
// is published on the error topic when an attempt to change the state document fails.
//
// It is only emitted as a response to a publish request on one of the reserved `$aws` topics.
//
// See [Device Shadow service documents] and [Device Shadow error messages] in the AWS IoT Core documentation.
//
// [Device Shadow service documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html#device-shadow-example-error-json
// [Device Shadow error messages]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-error-messages.html
type ErrorResponse struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Timestamp   int64  `json:"timestamp"`
	ClientToken string `json:"clientToken"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}
