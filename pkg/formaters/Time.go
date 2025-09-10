package formaters

import (
	"fmt"
	"time"
)

func RoundAndFormatDuration(timestamp time.Time) string {
	if timestamp.IsZero() {
		return "never"
	}

	d := time.Since(timestamp)

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%dm%ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh", hours)
	} else {
		days := int(d.Hours()) / 24
		return fmt.Sprintf("%dd", days)
	}
}
