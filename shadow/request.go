package shadow

// Request represents an AWS IoT Core Device Shadow request state document,
// which is published to report state or to set desired state.
//
// See [Device Shadow service documents] in the AWS IoT Core documentation.
//
// [Device Shadow service documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html#device-shadow-example-request-json
type Request[T any] struct {
	State       DesiredReported[T] `json:"state,omitempty"`
	ClientToken string             `json:"clientToken,omitempty"`
	Version     int                `json:"version,omitempty"`
}
