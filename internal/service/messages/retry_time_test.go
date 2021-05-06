package messages

import (
	"testing"
	"time"

	"github.com/kagelui/notification/internal/testutil"
)

func Test_getRetryTime(t *testing.T) {
	now := time.Now()
	type args struct {
		curr       time.Time
		retryCount int
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "retry 1",
			args: args{
				curr:       now,
				retryCount: 0,
			},
			want: now.Add(time.Minute*15),
		},
		{
			name: "retry 2",
			args: args{
				curr:       now,
				retryCount: 1,
			},
			want: now.Add(time.Minute*45),
		},
		{
			name: "retry 3",
			args: args{
				curr:       now,
				retryCount: 2,
			},
			want: now.Add(time.Minute*120),
		},
		{
			name: "retry 4",
			args: args{
				curr:       now,
				retryCount: 3,
			},
			want: now.Add(time.Minute*180),
		},
		{
			name: "retry 5",
			args: args{
				curr:       now,
				retryCount: 4,
			},
			want: now.Add(time.Minute*360),
		},
		{
			name: "retry 6",
			args: args{
				curr:       now,
				retryCount: 5,
			},
			want: now.Add(time.Minute*720),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.Equals(t, tt.want, getRetryTime(tt.args.curr, tt.args.retryCount))
		})
	}
}
