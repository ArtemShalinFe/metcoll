// Package build is used for build versioning.
package build

import (
	"fmt"
)

var buildVersion string
var buildDate string
var buildCommit string

// Info returns data about the assembly.
// Output format: "Build version: %s;\nBuild date: %s;\nBuild commit: %s;"
func Info() string {
	v := "N/A"
	if buildVersion != "" {
		v = buildVersion
	}

	d := "N/A"
	if buildDate != "" {
		d = buildDate
	}

	c := "N/A"
	if buildCommit != "" {
		c = buildCommit
	}

	return fmt.Sprintf("Build version: %s;\nBuild date: %s;\nBuild commit: %s;", v, d, c)
}
