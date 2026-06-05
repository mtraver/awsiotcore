package shadow

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMetadataUnmarshal(t *testing.T) {
	int64Ptr := func(v int64) *int64 { return &v }

	cases := []struct {
		name    string
		input   string
		want    Metadata
		wantErr bool
	}{
		{
			name:  "null",
			input: `null`,
			want:  Metadata{},
		},
		{
			name:  "timestamp leaf",
			input: `{"timestamp": 1780542138}`,
			want:  Metadata{Timestamp: int64Ptr(1780542138)},
		},
		{
			name:  "flat object",
			input: `{"attr1": {"timestamp": 1}, "attr2": {"timestamp": 2}}`,
			want: Metadata{
				Children: map[string]*Metadata{
					"attr1": {Timestamp: int64Ptr(1)},
					"attr2": {Timestamp: int64Ptr(2)},
				},
			},
		},
		{
			name:  "nested object",
			input: `{"attr1": {"attr2": {"timestamp": 1}}}`,
			want: Metadata{
				Children: map[string]*Metadata{
					"attr1": {
						Children: map[string]*Metadata{
							"attr2": {Timestamp: int64Ptr(1)},
						},
					},
				},
			},
		},
		{
			name:  "array of timestamp leaves",
			input: `[{"timestamp": 1}, {"timestamp": 2}]`,
			want: Metadata{
				Items: []*Metadata{
					{Timestamp: int64Ptr(1)},
					{Timestamp: int64Ptr(2)},
				},
			},
		},
		{
			name:  "array of objects",
			input: `[{"attr1": {"timestamp": 1}}, {"attr2": {"timestamp": 2}}]`,
			want: Metadata{
				Items: []*Metadata{
					{
						Children: map[string]*Metadata{
							"attr1": {Timestamp: int64Ptr(1)},
						},
					},
					{
						Children: map[string]*Metadata{
							"attr2": {Timestamp: int64Ptr(2)},
						},
					},
				},
			},
		},
		{
			name:  "field named timestamp",
			input: `{"timestamp": {"nested": {"timestamp": 1}}}`,
			want: Metadata{
				Children: map[string]*Metadata{
					"timestamp": {
						Children: map[string]*Metadata{
							"nested": {Timestamp: int64Ptr(1)},
						},
					},
				},
			},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got Metadata
			err := json.Unmarshal([]byte(tc.input), &got)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
