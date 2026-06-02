// package awsiotcore eases interaction with AWS IoT Core over MQTT.
// It handles TLS configuration and authentication.
package awsiotcore

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

// NewClient creates a github.com/eclipse/paho.mqtt.golang Client that may be used to connect to the device's MQTT broker using TLS.
// By default it sets up a github.com/eclipse/paho.mqtt.golang ClientOptions with the minimal
// options required to establish a connection:
//
//   - Broker
//   - Client ID set to the device's ID
//   - TLS configuration that supplies root CA certs, the device's cert, and Server Name Indication (SNI) (required by AWS IoT)
//
// By passing in options you may customize the ClientOptions. Options are functions with this signature:
//
//	func(*Device, *mqtt.ClientOptions) error
//
// They modify the ClientOptions. The option functions are applied to the ClientOptions in the order given before the
// Client is created. For example, if you wish to set the connect timeout, you might write this:
//
//	func ConnectTimeout(t time.Duration) func(*Device, *mqtt.ClientOptions) error {
//		return func(d *Device, opts *mqtt.ClientOptions) error {
//			opts.SetConnectTimeout(t)
//			return nil
//		}
//	}
//
// No options are required to establish a connection.
//
// For more information about connecting to AWS IoT MQTT brokers see https://docs.aws.amazon.com/iot/latest/developerguide/iot-connect-devices.html.
func (d *Device) NewClient(options ...func(*Device, *mqtt.ClientOptions) error) (mqtt.Client, error) {
	tlsConf := &tls.Config{
		RootCAs:      d.CACerts,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{d.Cert},
		// AWS IoT requires devices to send the Server Name Indication (SNI) TLS extension, and its value must be the endpoint address.
		// See https://docs.aws.amazon.com/iot/latest/developerguide/transport-security.html.
		ServerName: d.Endpoint,
		MinVersion: tls.VersionTLS12,
	}

	broker := d.Broker()

	// See https://docs.aws.amazon.com/iot/latest/developerguide/transport-security.html
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker.URL())
	opts.SetClientID(d.DeviceID)
	opts.SetTLSConfig(tlsConf)

	for _, option := range options {
		if err := option(d, opts); err != nil {
			return nil, err
		}
	}

	return mqtt.NewClient(opts), nil
}

func (d *Device) Broker() MQTTBroker {
	return MQTTBroker{
		Host: d.Endpoint,
		Port: 8883,
	}
}

func (d *Device) ID() string {
	return d.DeviceID
}

// TelemetryTopic returns the MQTT topic to which the device should publish telemetry events.
func (d *Device) TelemetryTopic() string {
	if d.TelemetryTopicOverride != "" {
		return d.TelemetryTopicOverride
	}
	return fmt.Sprintf("things/%v/telemetry", d.DeviceID)
}
