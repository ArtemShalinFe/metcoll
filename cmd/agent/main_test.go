package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_isTimeToPushReport(t *testing.T) {

	now := time.Now()
	lastReportPush = now

	have := isTimeToPushReport(now)
	assert.Equal(t, have, false)

	lastReportPush = time.Now().Add(-1 * time.Hour)
	have = isTimeToPushReport(now)
	assert.Equal(t, have, true)

}
