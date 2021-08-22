package jobs

import "errors"

var (
	ErrJobNotFound      = errors.New("job not found")
	ErrNoRowUpdated     = errors.New("no row updated")
	ErrJobWasClaimed    = errors.New("job was claimed")
	ErrJobWasNotClaimed = errors.New("job was not claimed")
	ErrJobExceedTimeout = errors.New("job exceed timeout")
)
