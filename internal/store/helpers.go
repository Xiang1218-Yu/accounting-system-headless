package store

import (
	"fmt"
	"time"
)

func formatPeriod(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

func defaultYear() int  { return time.Now().Year() }
func defaultMonth() int { return int(time.Now().Month()) }
