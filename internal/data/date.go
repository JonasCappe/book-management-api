package data

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

type Date struct {
	time.Time
}

func ParseDate(value string) (Date, error) {
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return Date{}, fmt.Errorf("date must use YYYY-MM-DD format: %w", err)
	}
	return Date{Time: parsed}, nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateLayout) + `"`), nil
}

func (d *Date) UnmarshalJSON(value []byte) error {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return fmt.Errorf("date must be a string in YYYY-MM-DD format")
	}
	parsed, err := ParseDate(string(value[1 : len(value)-1]))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d *Date) Scan(value any) error {
	parsed, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan %T into Date", value)
	}
	d.Time = parsed
	return nil
}

func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Format(dateLayout), nil
}
