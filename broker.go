package awsiotcore

import "fmt"

// MQTTBroker represents an MQTT server.
type MQTTBroker struct {
	Host string
	Port int
}

// URL returns the URL of the MQTT server.
func (b *MQTTBroker) URL() string {
	return fmt.Sprintf("tls://%s:%d", b.Host, b.Port)
}

// String returns a string representation of the MQTTBroker.
func (b *MQTTBroker) String() string {
	return b.URL()
}
