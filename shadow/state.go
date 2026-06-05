package shadow

// DesiredReported represents a pair of desired and reported states and is used
// in various AWS IoT Core Device Shadow requests and responses.
//
// T is the state that will be managed by the AWS IoT Core Device Shadow service.
// It must be marshallable into a JSON object.
type DesiredReported[T any] struct {
	Desired  T `json:"desired,omitempty"`
	Reported T `json:"reported,omitempty"`
}

// Desired represents a desired state and is used in various AWS IoT Core Device
// Shadow requests and responses.
//
// T is the state that will be managed by the AWS IoT Core Device Shadow service.
// It must be marshallable into a JSON object.
type Desired[T any] struct {
	Desired T `json:"desired,omitempty"`
}
