package awsiotcore

import (
	"testing"
)

func TestID(t *testing.T) {
	device := Device{
		Endpoint:    "myendpoint",
		DeviceID:    "foo",
		CertPath:    "foo.x509",
		PrivKeyPath: "foo.pem",
	}

	want := device.DeviceID
	got := device.ID()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTelemetryTopic(t *testing.T) {
	cases := []struct {
		name   string
		device Device
		want   string
	}{
		{
			name: "default_telemetry_topic",
			device: Device{
				Endpoint:    "myendpoint",
				DeviceID:    "foo",
				CertPath:    "foo.x509",
				PrivKeyPath: "foo.pem",
			},
			want: "things/foo/telemetry",
		},
		{
			name: "telemetry_topic_override",
			device: Device{
				Endpoint:               "myendpoint",
				DeviceID:               "foo",
				TelemetryTopicOverride: "things/foo/my/custom/topic",
				CertPath:               "foo.x509",
				PrivKeyPath:            "foo.pem",
			},
			want: "things/foo/my/custom/topic",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.device.TelemetryTopic()
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
