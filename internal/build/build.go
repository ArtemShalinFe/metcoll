// Package build is used for build versioning.
package build

import (
	"fmt"
)

var buildVersion string
var buildDate string
var buildCommit string

const infoTemplate = "Build version: %s; Build date: %s; Build commit: %s;"

// Info returns data about the assembly.
func Info() string {
	const NotAvailable = "N/A"

	v := NotAvailable
	if buildVersion != "" {
		v = buildVersion
	}

	d := NotAvailable
	if buildDate != "" {
		d = buildDate
	}

	c := NotAvailable
	if buildCommit != "" {
		c = buildCommit
	}

	return fmt.Sprintf(infoTemplate, v, d, c)
}
