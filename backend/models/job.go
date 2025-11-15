package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DateOnly is a custom type for date-only values (YYYY-MM-DD)
type DateOnly struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for DateOnly
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	// Remove quotes if present
	dateStr := string(data)
	if len(dateStr) >= 2 && dateStr[0] == '"' && dateStr[len(dateStr)-1] == '"' {
		dateStr = dateStr[1 : len(dateStr)-1]
	}
	
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %s, expected YYYY-MM-DD", dateStr)
	}
	
	d.Time = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

// MarshalJSON implements json.Marshaler for DateOnly
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2006-01-02"))
}

// Value implements driver.Valuer for database storage
func (d DateOnly) Value() (driver.Value, error) {
	return d.Time.Format("2006-01-02"), nil
}

// Scan implements sql.Scanner for database retrieval
func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	
	switch v := value.(type) {
	case time.Time:
		d.Time = time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC)
		return nil
	case string:
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		d.Time = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		return nil
	case []byte:
		return d.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan %T into DateOnly", value)
	}
}

type Job struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID        uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	Title            string     `gorm:"size:255;not null" json:"title"`
	Description      string     `gorm:"type:text;not null" json:"description"`
	Requirements     string     `gorm:"type:text" json:"requirements"`
	Location         string     `gorm:"size:255" json:"location"`
	JobType          string     `gorm:"size:50" json:"job_type"`
	SalaryRange      string     `gorm:"size:100" json:"salary_range"`
	Deadline         DateOnly   `gorm:"type:date;not null" json:"deadline"`
	Status           string     `gorm:"size:50;default:'open'" json:"status"`
	AutoShortlist    bool       `gorm:"default:true" json:"auto_shortlist"`
	ShortlistCriteria *string   `gorm:"type:jsonb" json:"shortlist_criteria,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relations
	Applications []Application `gorm:"foreignKey:JobID" json:"applications,omitempty"`
}

