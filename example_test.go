package awsiotcore_test

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"

	"github.com/mtraver/awsiotcore"
)

func Example() {
	// roots.pem should contain the root CA certs described in the README.
	caCerts, err := loadCACerts("roots.pem")
	if err != nil {
		log.Fatalf("Failed to load CA certs: %v", err)
	}

	// Load device certificate/private key pair.
	deviceCert, err := tls.LoadX509KeyPair("my-device.x509", "my-device.pem")
	if err != nil {
		log.Fatalf("Failed to load device cert/key pair: %v", err)
	}

	d := awsiotcore.Device{
		Endpoint: "my-endpoint",
		DeviceID: "my-device",
		CACerts:  caCerts,
		Cert:     deviceCert,
	}

	client, err := d.NewClient()
	if err != nil {
		log.Fatalf("Failed to make MQTT client: %v", err)
	}

	if token := client.Connect(); !token.Wait() || token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	if token := client.Publish(d.TelemetryTopic(), 1, false, []byte("{\"temp\": 18.0}")); !token.Wait() || token.Error() != nil {
		log.Printf("Failed to publish: %v", token.Error())
	}

	client.Disconnect(500)
}

func loadCACerts(path string) (*x509.CertPool, error) {
	pemCerts, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(pemCerts) {
		return nil, errors.New("no certs were parsed")
	}

	return certpool, nil
}
