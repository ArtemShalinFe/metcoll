package build

import (
	"fmt"
	"testing"
	"time"
)

func TestInfo(t *testing.T) {
	now := time.Now()
	date := now.Format("20060102150405")

	tests := []struct {
		name         string
		buildVersion string
		buildDate    string
		buildCommit  string
		want         string
	}{
		{
			name: "case #1",
			want: fmt.Sprintf(infoTemplate(), notAvailable, notAvailable, notAvailable),
		},
		{
			name:         "case #2",
			buildVersion: "1.0",
			buildDate:    date,
			buildCommit:  "test",
			want:         fmt.Sprintf(infoTemplate(), "1.0", date, "test"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buildDate = tt.buildDate
			buildCommit = tt.buildCommit
			buildVersion = tt.buildVersion

			b := NewBuild()
			if got := b.String(); got != tt.want {
				t.Errorf("b.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
