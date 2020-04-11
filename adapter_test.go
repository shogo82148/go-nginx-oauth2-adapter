package adapter

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	String := func(s string) *string {
		return &s
	}
	dummyProvider := map[string]map[string]interface{}{"development": {}}

	testcases := []struct {
		err bool
		c   Config
	}{
		// fail with "provider configure not found" error.
		{
			err: true,
			c:   Config{},
		},

		// test "the session authentication key is empty" error.
		{
			err: false, // just warn in running.
			c: Config{
				Providers: dummyProvider,
			},
		},
		{
			err: true, // fail in testing the config.
			c: Config{
				Providers:  dummyProvider,
				ConfigTest: true,
			},
		},

		// check session authentication key length.
		{
			err: false, // the secret is 32 bytes, and it's valid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("dummy-session-authentication-key")},
			},
		},
		{
			err: true, // the secret is 33 bytes, and it's invalid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("dummy-session-authentication-key+")},
			},
		},
		{
			err: false, // valid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("8e26ea01bd8805788bcb4660c7c15692e4771b5d6a22635eede025ca544ad4a00bcd17295f1ca8a5d573899fc7a641a25f488c9a5e839368cd79c2ffe1028031")},
			},
		},
		{
			err: true, // invalid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("INVALID-HEX-05788bcb4660c7c15692e4771b5d6a22635eede025ca544ad4a00bcd17295f1ca8a5d573899fc7a641a25f488c9a5e839368cd79c2ffe1028031")},
			},
		},

		// check session encryption key length.
		{
			err: false, // the secret is 32 bytes, and it's valid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{nil, String("**dummy-session-encryption-key**")},
			},
		},
		{
			err: true, // the secret is 33 bytes, and it's invalid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{nil, String("dummy-session-encryption-key")},
			},
		},
		{
			err: false, // valid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{nil, String("5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea86404921658e")},
			},
		},
		{
			err: true, // invalid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{nil, String("INVALID-HEX-5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea")},
			},
		},

		// invalid duration.
		{
			err: true,
			c: Config{
				Providers:          dummyProvider,
				AppRefreshInterval: "hoge",
			},
		},
	}

	for i, tc := range testcases {
		s, err := NewServer(tc.c)
		t.Logf("%d %#v: %v, %v", i, tc.c, s, err)
		if tc.err {
			if err == nil {
				t.Errorf("%v: expected error, got no error", tc.c)
			}
		} else {
			if err != nil {
				t.Errorf("%v: expected no error, got %v", tc.c, err)
			}
		}
		if err != nil {
			continue
		}
	}
}
