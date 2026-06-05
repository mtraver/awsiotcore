package shadow

import (
	"encoding/json"
	"fmt"
)

// Metadata represents a node in the metadata tree that is included in AWS IoT
// Core Device Shadow service responses.
//
// It contains the timestamp at which each attribute in the state document was updated.
// It follows the structure of the state document. Each node is one of a timestamp leaf,
// an object node, an array node, or null.
//
// See [Device Shadow service documents] in the AWS IoT Core documentation.
//
// [Device Shadow service documents]: https://docs.aws.amazon.com/iot/latest/developerguide/device-shadow-document.html
type Metadata struct {
	Timestamp *int64
	Children  map[string]*Metadata
	Items     []*Metadata
}

func (m *Metadata) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// Try unmarshaling as an array.
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		m.Items = make([]*Metadata, len(arr))
		for i, v := range arr {
			var child Metadata
			if err := json.Unmarshal(v, &child); err != nil {
				return err
			}
			m.Items[i] = &child
		}

		return nil
	}

	// Try unmarshaling as an object.
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err == nil {
		for k, v := range obj {
			if k == "timestamp" {
				// Failure to unmarshal as an int means that this is a field
				// called "timestamp", not a metadata timestamp, which is fine.
				var ts int64
				if err := json.Unmarshal(v, &ts); err == nil {
					m.Timestamp = &ts
					continue
				}
			}

			if m.Children == nil {
				m.Children = make(map[string]*Metadata)
			}

			var child Metadata
			if err := json.Unmarshal(v, &child); err != nil {
				return err
			}
			m.Children[k] = &child
		}

		return nil
	}

	return fmt.Errorf("metadata: unexpected JSON value: %q", data)
}

type MetadataDesired struct {
	Desired Metadata `json:"desired"`
}

type MetadataDesiredReported struct {
	Desired  Metadata `json:"desired"`
	Reported Metadata `json:"reported"`
}
