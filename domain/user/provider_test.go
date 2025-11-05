package user

import "testing"

func TestProvider_isValid(t *testing.T) {
	tests := []struct {
		name string
		p    Provider
		want error
	}{
		{name: "anonymous", p: ProviderAnonymous, want: nil},
		{name: "google", p: ProviderGoogle, want: nil},
		{name: "facebook", p: ProviderFacebook, want: nil},
		{name: "invalid", p: Provider("no"), want: ErrInvalidProvider(Provider("no"))},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.p.isValid()
			if tt.want != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.want)
				}
				if err.Error() != tt.want.Error() {
					t.Fatalf("error mismatch: want %v got %v", tt.want, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
