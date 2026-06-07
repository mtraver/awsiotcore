// package awsiotcore eases interaction with AWS IoT Core over MQTT.
// It handles TLS configuration and authentication.
package awsiotcore

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

// Device represents an AWS IoT device.
type Device struct {
	Endpoint               string
	DeviceID               string
	TelemetryTopicOverride string

	// CACerts must contain Amazon's trusted root certs. See the README for more info.
	CACerts *x509.CertPool

	// Cert if the device's certificate.
	Cert tls.Certificate
}

// ClientConfig creates an [autopaho.ClientConfig] that may be used to connect
// to the device's MQTT broker using TLS.
//
// The returned [autopaho.ClientConfig] includes the minimal required configuration
// to establish a connection:
//
//   - Broker
//   - Client ID set to the device's ID
//   - TLS configuration that supplies root CA certs, the device's cert, and
//     Server Name Indication (SNI) (required by AWS IoT)
//
// The caller is responsible for adding any additional configuration such as
// connection and message handlers.
//
// For more information about connecting to AWS IoT MQTT brokers see the documentation [here].
//
// [here]: https://docs.aws.amazon.com/iot/latest/developerguide/iot-connect-devices.html
func (d *Device) ClientConfig() autopaho.ClientConfig {
	tlsConf := &tls.Config{
		RootCAs:      d.CACerts,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{d.Cert},
		// AWS IoT requires devices to send the Server Name Indication (SNI) TLS extension,
		// and its value must be the endpoint address.
		// See https://docs.aws.amazon.com/iot/latest/developerguide/transport-security.html.
		ServerName: d.Endpoint,
		MinVersion: tls.VersionTLS12,
	}

	// See https://docs.aws.amazon.com/iot/latest/developerguide/transport-security.html
	cfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{d.Broker().URL()},
		TlsCfg:     tlsConf,
		ClientConfig: paho.ClientConfig{
			ClientID: d.ID(),
		},
	}

	return cfg
}

func (d *Device) Broker() *MQTTBroker {
	return &MQTTBroker{
		Host: d.Endpoint,
		Port: 8883,
	}
}

func (d *Device) ID() string {
	return d.DeviceID
}

// TelemetryTopic returns the MQTT topic to which the device should publish telemetry.
func (d *Device) TelemetryTopic() string {
	if d.TelemetryTopicOverride != "" {
		return d.TelemetryTopicOverride
	}
	return fmt.Sprintf("things/%s/telemetry", d.DeviceID)
}
