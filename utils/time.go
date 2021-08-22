package utils

import "time"

func TimeToPtr(t time.Time) *time.Time {
	return &t
}

func TimeNow() time.Time {
	return time.Now().UTC()
}
