package shadow

// AcceptedResponse represents responses published to AWS IoT Core's reserved [/accepted topics].
//
// See [Device Shadow service documents] in the AWS IoT Core documentation.
//
// [/accepted topics]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-mqtt.html
// [Device Shadow service documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html#device-shadow-example-response-json
type AcceptedResponse[T any] struct {
	State       Desired[T]      `json:"state"`
	Metadata    MetadataDesired `json:"metadata"`
	Timestamp   int64           `json:"timestamp"`
	ClientToken string          `json:"clientToken"`
	Version     int             `json:"version"`
}

// DeltaResponse represents the response that is published to AWS IoT Core's reserved
// [/update/delta topic] when a change to the device's shadow is accepted, and the
// response state document contains different values for `desired` and `reported` states.
//
// A DeltaResponse may be handled via an implementation of the [UpdateHandler] interface
// provided to [NewClient].
//
// The information below is copied from [Delta state] in the AWS IoT Core documentation.
// Also see [Device Shadow service documents].
//
// Delta state is a virtual type of state that contains the difference between
// the `desired` and `reported` states. Fields in the `desired` section that are
// not in the `reported` section are included in the delta. Fields that are in the
// `reported` section and not in the `desired` section are not included in the
// delta. The delta contains metadata, and its values are equal to the metadata in
// the `desired` field.
//
// When nested objects differ, the delta contains the path all the way to the root.
//
// The Device Shadow service calculates the delta by iterating through each field in
// the `desired` state and comparing it to the `reported` state.
//
// Arrays are treated like values. If an array in the `desired` section doesn't match
// the array in the `reported` section, then the entire `desired` array is copied into
// the delta.
//
// [/update/delta topic]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-mqtt.html#update-delta-pub-sub-topic
// [Delta state]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html#delta-state
// [Device Shadow service documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html#device-shadow-example-response-json
type DeltaResponse[T any] struct {
	State       T        `json:"state"`
	Metadata    Metadata `json:"metadata"`
	Timestamp   int64    `json:"timestamp"`
	ClientToken string   `json:"clientToken"`
	Version     int      `json:"version"`
}

// DocumentsResponse represents the state document that is published to AWS IoT Core's reserved
// [/update/documents topic] whenever an update to the shadow is successfully performed.
//
// A DocumentsResponse may be handled via an implementation of the [UpdateHandler] interface
// provided to [NewClient].
//
// See [Device Shadow service documents] in the AWS IoT Core documentation.
//
// [/update/documents topic]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-mqtt.html#update-documents-pub-sub-topic
// [Device Shadow service documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html#device-shadow-example-response-json
type DocumentsResponse[T any] struct {
	Previous    StateSnapshot[T] `json:"previous"`
	Current     StateSnapshot[T] `json:"current"`
	Timestamp   int64            `json:"timestamp"`
	ClientToken string           `json:"clientToken"`
}

type StateSnapshot[T any] struct {
	State    DesiredReported[T]      `json:"state"`
	Metadata MetadataDesiredReported `json:"metadata"`
	Version  int                     `json:"version"`
}
