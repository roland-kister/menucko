package util

import "time"

type DateResolver interface {
	Weekday() int
}

type DevDateResolver struct {
	WeekdayVal int
}

func (resolver DevDateResolver) Weekday() int {
	return resolver.WeekdayVal
}

type ProdDateResolver struct{}

func (ProdDateResolver) Weekday() int {
	weekday := time.Now().UTC().Weekday()

	if weekday == 0 {
		weekday = 6
	} else {
		weekday -= 1
	}

	return int(weekday)
}
