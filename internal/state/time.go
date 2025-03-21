package state

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TimeValue struct {
	Hour   int
	Minute int
	Second int
	Valid  bool
}

func (t TimeValue) Validate() error {
	if t.Hour < 0 || t.Hour > 23 {
		return errors.New("hour must be between 0 and 23")
	}
	if t.Minute < 0 || t.Minute > 59 {
		return errors.New("minute must be between 0 and 59")
	}
	if t.Second < 0 || t.Second > 59 {
		return errors.New("second must be between 0 and 59")
	}
	return nil
}

func (t TimeValue) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

func (t *TimeValue) Scan(value interface{}) error {
	if value == nil {
		*t = TimeValue{Valid: false}
		return nil
	}

	var timeStr string

	switch v := value.(type) {
	case []byte:
		timeStr = string(v)
	case string:
		timeStr = v
	case time.Time:

		*t = TimeValue{
			Hour:   v.Hour(),
			Minute: v.Minute(),
			Second: v.Second(),
			Valid:  true,
		}
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing %T into TimeValue", value)
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid time format: %s, expected HH:MM:SS", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid minute: %s", parts[1])
	}

	second, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid second: %s", parts[2])
	}

	*t = TimeValue{
		Hour:   hour,
		Minute: minute,
		Second: second,
		Valid:  true,
	}

	if err := t.Validate(); err != nil {
		return err
	}

	return nil
}

func (t TimeValue) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t.String(), nil
}
