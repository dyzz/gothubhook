package main

import (
	"fmt"
	"time"
)

func ThisWeek() string {
	date := time.Now()
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	end := date.AddDate(0, 0, 6)
	return fmt.Sprintf("%s - %s", date.Format("01/02"), end.Format("01/02"))
}
