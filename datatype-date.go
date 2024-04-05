package formx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Date struct {
	Year  int32
	Month int32
	Day   int32
}

func NewDateFromTime(time time.Time) Date {
	return Date{
		Year:  int32(time.Year()),
		Month: int32(time.Month()),
		Day:   int32(time.Day()),
	}
}

func (d Date) String() string {
	if d.Year == 0 {
		return ""
	}
	return fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) Time() time.Time {
	return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), 0, 0, 0, 0, time.UTC)
}

func (d *Date) ParseString(s string) error {
	if s == "" {
		d.Year = 0
		d.Month = 0
		d.Day = 0
		return nil
	}
	ss := strings.Split(s, "-")
	if len(ss) != 3 {
		return fmt.Errorf("invalid date format")
	}

	year, err := strconv.ParseInt(ss[0], 10, 32)
	if err != nil {
		return err
	}

	month, err := strconv.ParseInt(ss[1], 10, 32)
	if err != nil {
		return err
	}

	day, err := strconv.ParseInt(ss[2], 10, 32)
	if err != nil {
		return err
	}

	d.Year = int32(year)
	d.Month = int32(month)
	d.Day = int32(day)

	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	return d.ParseString(s)
}
