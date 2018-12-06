package intervals

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

const second = 1000 * 1000 * 1000
const minute = second * 60
const hour = minute * 60
const day = hour * 24
const week = day * 7
const month = day * 30

func nanoseconds(unit string) int64 {
	switch unit {
	case "h":
		return hour
	case "d":
		return day
	case "w":
		return week
	case "M":
		return month
	case "m":
		return minute
	}
	panic("Unsupported unit")
}

var re = regexp.MustCompile("^(\\d+)([wdmMh])$")

func Validate(interval string) error {
	if !re.MatchString(interval) {
		return errors.New("Invalid interval")
	}
	return nil
}

func MustParse(interval string) time.Duration {
	d, err := Parse(interval)
	if err != nil {
		panic(err)
	}
	return d
}

func Parse(interval string) (time.Duration, error) {
	err := Validate(interval)
	if err != nil {
		return 0, err
	}
	tokens := re.FindStringSubmatch(interval)
	value, err := strconv.ParseInt(tokens[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(value * nanoseconds(tokens[2])), nil
}
