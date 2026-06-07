package awsiotcore

import (
	"testing"
)

func TestID(t *testing.T) {
	device := Device{
		Endpoint: "myendpoint",
		DeviceID: "foo",
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
			name: "default telemetry topic",
			device: Device{
				Endpoint: "myendpoint",
				DeviceID: "foo",
			},
			want: "things/foo/telemetry",
		},
		{
			name: "telemetry topic override",
			device: Device{
				Endpoint:               "myendpoint",
				DeviceID:               "foo",
				TelemetryTopicOverride: "things/foo/my/custom/topic",
			},
			want: "things/foo/my/custom/topic",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.device.TelemetryTopic()
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
