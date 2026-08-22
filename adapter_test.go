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
		name string
		err  bool
		c    Config
	}{
		// fail with "provider configure not found" error.
		{
			name: "provider configure not found",
			err:  true,
			c:    Config{},
		},

		// test "the session authentication key is empty" error.
		{
			name: "the session authentication key is empty",
			err:  true,
			c: Config{
				Providers: dummyProvider,
			},
		},
		{
			name: "the session authentication key is empty and config test is true",
			err:  true, // fail in testing the config.
			c: Config{
				Providers:  dummyProvider,
				ConfigTest: true,
			},
		},

		// check session authentication key length.
		{
			name: "the session authentication key is 32 bytes",
			err:  false, // the secret is 32 bytes, and it's valid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("dummy-session-authentication-key")},
			},
		},
		{
			name: "the session authentication key is 33 bytes",
			err:  true, // the secret is 33 bytes, and it's invalid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("dummy-session-authentication-key+")},
			},
		},
		{
			name: "the session authentication key is valid hex string",
			err:  false, // valid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("8e26ea01bd8805788bcb4660c7c15692e4771b5d6a22635eede025ca544ad4a00bcd17295f1ca8a5d573899fc7a641a25f488c9a5e839368cd79c2ffe1028031")},
			},
		},
		{
			name: "the session authentication key is invalid hex string",
			err:  true, // invalid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("INVALID-HEX-05788bcb4660c7c15692e4771b5d6a22635eede025ca544ad4a00bcd17295f1ca8a5d573899fc7a641a25f488c9a5e839368cd79c2ffe1028031")},
			},
		},

		// check session encryption key length.
		{
			name: "the session encryption key is 32 bytes",
			err:  false, // the secret is 32 bytes, and it's valid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea86404921658e"), String("**dummy-session-encryption-key**")},
			},
		},
		{
			name: "the session encryption key is 33 bytes",
			err:  true, // the secret is 33 bytes, and it's invalid.
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea86404921658e"), String("dummy-session-encryption-key")},
			},
		},
		{
			name: "the session encryption key is valid hex string",
			err:  false, // valid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea86404921658e"), String("5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea86404921658e")},
			},
		},
		{
			name: "the session encryption key is invalid hex string",
			err:  true, // invalid hex string
			c: Config{
				Providers: dummyProvider,
				Secrets:   []*string{String("5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea86404921658e"), String("INVALID-HEX-5c9ea31b400099a521f934f8a4c2c88758ca59e0a34479775aea")},
			},
		},

		// invalid duration.
		{
			name: "invalid duration",
			err:  true,
			c: Config{
				Providers:          dummyProvider,
				AppRefreshInterval: "hoge",
			},
		},
	}

	for _, tc := range testcases {
		_, err := NewServer(tc.c)
		if tc.err {
			if err == nil {
				t.Errorf("%s: expected error, got no error", tc.name)
			}
		} else {
			if err != nil {
				t.Errorf("%s: expected no error, got %v", tc.name, err)
			}
		}
		if err != nil {
			continue
		}
	}
}
