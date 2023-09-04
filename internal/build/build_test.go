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
			want: "Build version: N/A;\nBuild date: N/A;\nBuild commit: N/A;",
		},
		{
			name:         "case #2",
			buildVersion: "1.0",
			buildDate:    date,
			buildCommit:  "test",
			want:         fmt.Sprintf("Build version: %s;\nBuild date: %s;\nBuild commit: %s;", "1.0", date, "test"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buildDate = tt.buildDate
			buildCommit = tt.buildCommit
			buildVersion = tt.buildVersion

			if got := Info(); got != tt.want {
				t.Errorf("Info() = %v, want %v", got, tt.want)
			}
		})
	}
}
