package awsiotcore_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
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

	clientConfig := d.ClientConfig()
	clientConfig.KeepAlive = 20
	clientConfig.SessionExpiryInterval = 120

	// We'll run until cancelled by the user (e.g. ctrl-c).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cm, err := autopaho.NewConnection(ctx, clientConfig)
	if err != nil {
		log.Fatalf("Failed to create MQTT connection: %v", err)
	}

	if err := cm.AwaitConnection(ctx); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if _, err = cm.Publish(ctx, &paho.Publish{
				Topic:   d.TelemetryTopic(),
				Payload: []byte(`{"temp": 18.0}`),
				QoS:     1,
			}); err != nil {
				if ctx.Err() == nil {
					log.Printf("Failed to publish: %v", err)
				}
			}

			continue

		case <-ctx.Done():
		}

		break
	}

	log.Println("Exiting...")
	<-cm.Done()
	log.Println("Done")
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
