package envvar

import (
	"os"
	"testing"
	"time"

	"github.com/kagelui/notification/internal/testutil"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		target  interface{}
		envVar  map[string]string
		want    interface{}
		wantErr string
	}{
		{
			name: "not a pointer",
			target: struct {
				Anything string
			}{},
			envVar:  nil,
			want:    nil,
			wantErr: "invalid type struct { Anything string } passed",
		},
		{
			name:    "nil passed",
			target:  nil,
			envVar:  nil,
			want:    nil,
			wantErr: "invalid type <nil> passed",
		},
		{
			name:    "not a struct pointer",
			target:  &[]string{},
			envVar:  nil,
			want:    nil,
			wantErr: "not a struct pointer",
		},
		{
			name: "env tag not set for some field",
			target: &struct {
				One      string `env:"one"`
				Anything string
			}{},
			envVar:  nil,
			want:    nil,
			wantErr: "env tag not set for field Anything",
		},
		{
			name: "some env var not set",
			target: &struct {
				One   string `env:"one"`
				Two   string `env:"-"`
				Three string `env:"three"`
				Four  string `env:"-"`
				Five  string `env:"five"`
			}{},
			envVar:  map[string]string{"one": "hey", "two": "yes", "four": "yeah", "five": "some"},
			want:    nil,
			wantErr: "three not present",
		},
		{
			name: "un-export something",
			target: &struct {
				One string `env:"one"`
				two string `env:"two"`
			}{},
			envVar:  map[string]string{"one": "hey", "two": "yes"},
			want:    nil,
			wantErr: "field two is not valid or cannot be set",
		},
		{
			name: "int overflow",
			target: &struct {
				One   string  `env:"one"`
				Two   string  `env:"-"`
				Three int     `env:"three"`
				Four  string  `env:"-"`
				Five  float64 `env:"five"`
			}{},
			envVar:  map[string]string{"one": "hey", "two": "yes", "three": "3198619363846193128369139826", "four": "yeah", "five": "861.8362"},
			want:    nil,
			wantErr: `parsing "3198619363846193128369139826": value out of range`,
		},
		{
			name: "float64 overflow",
			target: &struct {
				One   string  `env:"one"`
				Two   string  `env:"-"`
				Three int     `env:"three"`
				Four  string  `env:"-"`
				Five  float64 `env:"five"`
			}{},
			envVar:  map[string]string{"one": "hey", "two": "yes", "three": "312", "four": "yeah", "five": "2e+308"},
			want:    nil,
			wantErr: `parsing "2e+308": value out of range`,
		},
		{
			name: "invalid time duration",
			target: &struct {
				One   string        `env:"one"`
				Two   string        `env:"-"`
				Three int           `env:"three"`
				Four  string        `env:"-"`
				Five  time.Duration `env:"five"`
			}{},
			envVar:  map[string]string{"one": "hey", "two": "yes", "three": "319826", "four": "yeah", "five": "10seconds"},
			want:    nil,
			wantErr: `unknown unit "seconds" in duration "10seconds"`,
		},
		{
			name: "unsupported type",
			target: &struct {
				One   string     `env:"one"`
				Two   string     `env:"-"`
				Three int        `env:"three"`
				Four  string     `env:"-"`
				Five  complex128 `env:"five"`
			}{},
			envVar:  map[string]string{"one": "hey", "two": "yes", "three": "319826", "four": "yeah", "five": "10s"},
			want:    nil,
			wantErr: "unsupported type complex128",
		},
		{
			name: "everything set",
			target: &struct {
				One string `env:"one"`
				Two string `env:"two"`
			}{},
			envVar: map[string]string{"one": "hey", "two": "yes"},
			want: &struct {
				One string `env:"one"`
				Two string `env:"two"`
			}{One: "hey", Two: "yes"},
			wantErr: "",
		},
		{
			name: "skip some",
			target: &struct {
				One   string        `env:"one"`
				Two   string        `env:"-"`
				Three int           `env:"three"`
				Four  string        `env:"-"`
				Five  float64       `env:"five"`
				Six   time.Duration `env:"six"`
			}{},
			envVar: map[string]string{"one": "hey", "two": "yes", "three": "319826", "four": "yeah", "five": "861.8362", "six": "5s"},
			want: &struct {
				One   string        `env:"one"`
				Two   string        `env:"-"`
				Three int           `env:"three"`
				Four  string        `env:"-"`
				Five  float64       `env:"five"`
				Six   time.Duration `env:"six"`
			}{One: "hey", Two: "", Three: 319826, Four: "", Five: 861.8362, Six: time.Second * 5},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVar {
				testutil.Ok(t, os.Setenv(k, v))
			}

			err := Read(tt.target)
			testutil.CompareError(t, tt.wantErr, err)
			if err == nil {
				testutil.Equals(t, tt.want, tt.target)
			}

			for k := range tt.envVar {
				testutil.Ok(t, os.Unsetenv(k))
			}
		})
	}
}
