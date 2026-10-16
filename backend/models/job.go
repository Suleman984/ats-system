package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

// DateOnly stores a calendar date and serializes as YYYY-MM-DD.
type DateOnly struct {
	time.Time
}

func ParseDateOnly(s string) (DateOnly, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return DateOnly{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", s)
	}
	return DateOnly{t}, nil
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseDateOnly(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(dateLayout))
}

func (d DateOnly) Value() (driver.Value, error) {
	return d.Format(dateLayout), nil
}

func (d *DateOnly) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		d.Time = time.Time{}
	case time.Time:
		d.Time = time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC)
	case string:
		parsed, err := ParseDateOnly(v)
		if err != nil {
			return err
		}
		*d = parsed
	case []byte:
		return d.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan %T into DateOnly", value)
	}
	return nil
}

type Job struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID         uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Requirements      string    `json:"requirements"`
	Location          string    `json:"location"`
	JobType           string    `json:"job_type"`
	SalaryRange       string    `json:"salary_range"`
	Deadline          DateOnly  `gorm:"type:date" json:"deadline"`
	Status            string    `gorm:"default:open" json:"status"`
	AutoShortlist     bool      `json:"auto_shortlist"`
	ShortlistCriteria *string   `gorm:"type:jsonb" json:"shortlist_criteria,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Closed reports whether the job no longer accepts applications.
func (j *Job) Closed(now time.Time) bool {
	return j.Status != "open" || !now.Before(j.Deadline.AddDate(0, 0, 1))
}
