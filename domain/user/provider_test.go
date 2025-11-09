package user

import (
	"testing"

	"github.com/naka-sei/tsudzuri/pkg/testutil"
)

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
			testutil.EqualErr(t, tt.want, err)
		})
	}
}
