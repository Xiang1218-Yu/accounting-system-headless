package service

import "fmt"

func formatPeriod2(year, month int) string {
	return fmt.Sprintf("%04d年%02d月", year, month)
}
