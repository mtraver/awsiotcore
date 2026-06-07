package awsiotcore

import (
	"net"
	"net/url"
	"strconv"
)

// MQTTBroker represents an MQTT broker.
type MQTTBroker struct {
	Host string
	Port int
}

// URL returns the URL of the MQTT broker.
func (b *MQTTBroker) URL() *url.URL {
	return &url.URL{
		Scheme: "tls",
		Host:   net.JoinHostPort(b.Host, strconv.Itoa(b.Port)),
	}
}

func (b *MQTTBroker) String() string {
	return b.URL().String()
}
