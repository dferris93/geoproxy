package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckIPType(t *testing.T) {
	checker := &CheckIPs{}

	ipType, err := checker.CheckIPType("127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, 4, ipType)

	ipType, err = checker.CheckIPType("2001:db8::1")
	assert.NoError(t, err)
	assert.Equal(t, 6, ipType)

	_, err = checker.CheckIPType("not-an-ip")
	assert.Error(t, err)
}

func TestCheckSubnets(t *testing.T) {
	checker := &CheckIPs{}

	found := checker.CheckSubnets([]string{"127.0.0.1"}, "127.0.0.1")
	assert.True(t, found)

	found = checker.CheckSubnets([]string{"2001:db8::1"}, "2001:db8::1")
	assert.True(t, found)

	found = checker.CheckSubnets([]string{"bad", "10.0.0.0/24"}, "10.0.1.1")
	assert.False(t, found)
}

func TestMakeSetAndSubtractSet(t *testing.T) {
	set := MakeSet([]string{"a", "b", "b"})
	assert.True(t, set["a"])
	assert.True(t, set["b"])

	remaining := SubtractSet(set, map[string]bool{"b": true, "c": true})
	assert.True(t, remaining["a"])
	assert.False(t, remaining["b"])
}

func TestMakeNormalizedUpperSet(t *testing.T) {
	set := MakeNormalizedUpperSet([]string{" us ", "ca", "CA", "", "  "})
	assert.True(t, set["US"])
	assert.True(t, set["CA"])
	assert.False(t, set[""])
	assert.Len(t, set, 2)
}

func TestCheckTime(t *testing.T) {
	location := time.UTC
	start := time.Date(2024, time.January, 1, 9, 0, 0, 0, location)
	end := time.Date(2024, time.January, 1, 17, 0, 0, 0, location)

	ok, err := CheckTime(start, end, time.Date(2024, time.January, 1, 12, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckTime(start, end, time.Date(2024, time.January, 1, 18, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckTimeWrapAround(t *testing.T) {
	location := time.UTC
	start := time.Date(2024, time.January, 1, 22, 0, 0, 0, location)
	end := time.Date(2024, time.January, 2, 6, 0, 0, 0, location)

	ok, err := CheckTime(start, end, time.Date(2024, time.January, 1, 23, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckTime(start, end, time.Date(2024, time.January, 2, 5, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckTime(start, end, time.Date(2024, time.January, 2, 12, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckDateRange(t *testing.T) {
	location := time.UTC
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, location)
	end := time.Date(2024, time.January, 10, 0, 0, 0, 0, location)

	ok, err := CheckDateRange(start, end, time.Date(2024, time.January, 5, 12, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = CheckDateRange(start, end, time.Date(2023, time.December, 31, 12, 0, 0, 0, location))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckDateRangeInvalid(t *testing.T) {
	location := time.UTC
	start := time.Date(2024, time.January, 10, 0, 0, 0, 0, location)
	end := time.Date(2024, time.January, 1, 0, 0, 0, 0, location)

	_, err := CheckDateRange(start, end, time.Date(2024, time.January, 5, 12, 0, 0, 0, location))
	assert.Error(t, err)
}
