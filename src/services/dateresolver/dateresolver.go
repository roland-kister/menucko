package dateresolver

import "time"

type DateResolver interface {
	Weekday() int
	SlovakWeekday() string
}

type DevDateResolver struct {
	WeekdayVal int
}

func (resolver DevDateResolver) Weekday() int {
	return resolver.WeekdayVal
}

func (resolver DevDateResolver) SlovakWeekday() string {
	return slovakDays[resolver.Weekday()]
}

var slovakDays = [...]string{"Pondelok", "Utorok", "Streda", "Štvrtok", "Piatok", "Sobota", "Nedeľa"}

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

func (resolver ProdDateResolver) SlovakWeekday() string {
	return slovakDays[resolver.Weekday()]
}
