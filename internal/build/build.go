// Package build is used for build versioning.
package build

import (
	"fmt"
)

var buildVersion string
var buildDate string
var buildCommit string

const notAvailable = "N/A"

type Build struct {
	buildVersion string
	buildDate    string
	buildCommit  string
}

// NewBuild - Object constructor.
func NewBuild() Build {
	b := Build{
		buildVersion: notAvailable,
		buildDate:    notAvailable,
		buildCommit:  notAvailable,
	}

	if buildVersion != "" {
		b.buildVersion = buildVersion
	}

	if buildDate != "" {
		b.buildDate = buildDate
	}

	if buildCommit != "" {
		b.buildCommit = buildCommit
	}

	return b
}

// Info returns data about the assembly.
func (b *Build) String() string {
	return fmt.Sprintf(infoTemplate(), b.buildVersion, b.buildDate, b.buildCommit)
}

func infoTemplate() string {
	return "Build version: %s; Build date: %s; Build commit: %s;"
}
