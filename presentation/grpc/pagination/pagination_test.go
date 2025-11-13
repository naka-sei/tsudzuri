package pagination

import (
	"errors"
	"testing"
)

func ptr(v int32) *int32 {
	return &v
}

func TestValidtePage(t *testing.T) {
	tests := []struct {
		name    string
		input   *int32
		want    int32
		wantErr error
	}{
		{
			name:  "nil_defaults_to_one",
			input: nil,
			want:  defaultPage,
		},
		{
			name:  "valid_page",
			input: ptr(3),
			want:  3,
		},
		{
			name:    "below_minimum",
			input:   ptr(0),
			wantErr: ErrInvalidPage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidtePage(tt.input)
			if got != tt.want {
				t.Fatalf("ValidtePage() got=%d want=%d", got, tt.want)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidatePageSize(t *testing.T) {
	tests := []struct {
		name    string
		input   *int32
		want    int32
		wantErr error
	}{
		{
			name:  "nil_defaults",
			input: nil,
			want:  DefaultPageSize,
		},
		{
			name:  "valid_size",
			input: ptr(50),
			want:  50,
		},
		{
			name:    "below_min",
			input:   ptr(0),
			wantErr: ErrInvalidPage,
		},
		{
			name:    "above_max",
			input:   ptr(MaxPageSize + 1),
			wantErr: ErrInvalidPage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePageSize(tt.input)
			if got != tt.want {
				t.Fatalf("ValidatePageSize() got=%d want=%d", got, tt.want)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
