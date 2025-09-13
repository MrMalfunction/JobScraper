package types

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type JobDetails struct {
	JUlid          ulid.ULID
	JobID          string
	JobRole        string
	JobDetails     string
	JobPostDate    time.Time
	JobLink        string
	JobCompanyName string
}
