package schedulermgr

import (
	"fmt"
	"strings"
	"time"
)

type cronField struct {
	min        int
	max        int
	allowed    []int
	allowedAll bool
}

type cronExpression struct {
	second   *cronField
	minute   *cronField
	hour     *cronField
	day      *cronField
	month    *cronField
	weekday  *cronField
	timezone *time.Location
}

var (
	monthNames = map[string]int{
		"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
		"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
	}
	weekdayNames = map[string]int{
		"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
	}
)

func parseCrontab(expr string, timezoneStr string) (*cronExpression, error) {
	parts := strings.Split(expr, " ")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid crontab expression: expected 6 fields, got %d", len(parts))
	}

	second, err := parseField(parts[0], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid second field: %w", err)
	}

	minute, err := parseField(parts[1], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid minute field: %w", err)
	}

	hour, err := parseField(parts[2], 0, 23, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid hour field: %w", err)
	}

	day, err := parseField(parts[3], 1, 31, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid day field: %w", err)
	}

	month, err := parseField(parts[4], 1, 12, monthNames)
	if err != nil {
		return nil, fmt.Errorf("invalid month field: %w", err)
	}

	weekday, err := parseField(parts[5], 0, 6, weekdayNames)
	if err != nil {
		return nil, fmt.Errorf("invalid weekday field: %w", err)
	}

	var timezone *time.Location
	if timezoneStr == "" {
		timezone = time.Local
	} else {
		timezone, err = time.LoadLocation(timezoneStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
	}

	return &cronExpression{
		second:   second,
		minute:   minute,
		hour:     hour,
		day:      day,
		month:    month,
		weekday:  weekday,
		timezone: timezone,
	}, nil
}

func parseField(field string, min, max int, nameMap map[string]int) (*cronField, error) {
	result := &cronField{
		min:     min,
		max:     max,
		allowed: make([]int, 0),
	}

	if field == "*" {
		result.allowedAll = true
		return result, nil
	}

	if field == "?" {
		result.allowedAll = false
		return result, nil
	}

	parts := strings.Split(field, ",")
	for _, part := range parts {
		if strings.Contains(part, "/") {
			if err := parseStep(result, part); err != nil {
				return nil, err
			}
		} else if strings.Contains(part, "-") {
			if err := parseRange(result, part, nameMap); err != nil {
				return nil, err
			}
		} else {
			value, err := parseValue(part, nameMap)
			if err != nil {
				return nil, err
			}
			if value < min || value > max {
				return nil, fmt.Errorf("value %d out of range [%d, %d]", value, min, max)
			}
			result.allowed = append(result.allowed, value)
		}
	}

	return result, nil
}

func parseStep(cf *cronField, part string) error {
	rangePart := strings.Split(part, "/")
	if len(rangePart) != 2 {
		return fmt.Errorf("invalid step format: %s", part)
	}

	var start, end int
	var useAll bool

	if rangePart[0] == "*" {
		start = cf.min
		end = cf.max
		useAll = true
	} else if strings.Contains(rangePart[0], "-") {
		rangeValues := strings.Split(rangePart[0], "-")
		if len(rangeValues) != 2 {
			return fmt.Errorf("invalid range format: %s", rangePart[0])
		}
		var err error
		start, err = parseValue(rangeValues[0], nil)
		if err != nil {
			return err
		}
		end, err = parseValue(rangeValues[1], nil)
		if err != nil {
			return err
		}
	} else {
		var err error
		start, err = parseValue(rangePart[0], nil)
		if err != nil {
			return err
		}
		end = cf.max
		useAll = true
	}

	if start < cf.min || start > cf.max {
		return fmt.Errorf("start value %d out of range [%d, %d]", start, cf.min, cf.max)
	}
	if end < cf.min || end > cf.max {
		return fmt.Errorf("end value %d out of range [%d, %d]", end, cf.min, cf.max)
	}

	step, err := parseValue(rangePart[1], nil)
	if err != nil {
		return err
	}
	if step <= 0 {
		return fmt.Errorf("step must be positive, got %d", step)
	}

	if useAll {
		cf.allowedAll = true
	} else {
		for v := start; v <= end; v += step {
			cf.allowed = append(cf.allowed, v)
		}
	}

	return nil
}

func parseRange(cf *cronField, part string, nameMap map[string]int) error {
	rangeValues := strings.Split(part, "-")
	if len(rangeValues) != 2 {
		return fmt.Errorf("invalid range format: %s", part)
	}

	start, err := parseValue(rangeValues[0], nameMap)
	if err != nil {
		return err
	}
	end, err := parseValue(rangeValues[1], nameMap)
	if err != nil {
		return err
	}

	if start < cf.min || start > cf.max {
		return fmt.Errorf("start value %d out of range [%d, %d]", start, cf.min, cf.max)
	}
	if end < cf.min || end > cf.max {
		return fmt.Errorf("end value %d out of range [%d, %d]", end, cf.min, cf.max)
	}

	for v := start; v <= end; v++ {
		cf.allowed = append(cf.allowed, v)
	}

	return nil
}

func parseValue(value string, nameMap map[string]int) (int, error) {
	value = strings.TrimSpace(value)
	if nameMap != nil {
		if v, ok := nameMap[strings.ToUpper(value)]; ok {
			return v, nil
		}
	}
	var v int
	_, err := fmt.Sscanf(value, "%d", &v)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %s", value)
	}
	return v, nil
}

func (ce *cronExpression) getNextExecutionTime(from time.Time) time.Time {
	t := from.In(ce.timezone).Add(1 * time.Second)
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, ce.timezone)

	maxIterations := 4 * 365 * 24 * 60 * 60
	for i := 0; i < maxIterations; i++ {
		if ce.matches(t) {
			return t
		}
		t = t.Add(1 * time.Second)
	}

	return time.Time{}
}

func (ce *cronExpression) matches(t time.Time) bool {
	return ce.matchesField(ce.second, t.Second()) &&
		ce.matchesField(ce.minute, t.Minute()) &&
		ce.matchesField(ce.hour, t.Hour()) &&
		ce.matchesDayMonth(t) &&
		ce.matchesField(ce.month, int(t.Month()))
}

func (ce *cronExpression) matchesField(cf *cronField, value int) bool {
	if cf == nil {
		return true
	}
	if cf.allowedAll {
		return value >= cf.min && value <= cf.max
	}
	if len(cf.allowed) == 0 {
		return true
	}
	for _, v := range cf.allowed {
		if v == value {
			return true
		}
	}
	return false
}

func (ce *cronExpression) matchesDayMonth(t time.Time) bool {
	day := t.Day()
	weekday := int(t.Weekday())

	dayMatch := (len(ce.day.allowed) == 0 && !ce.day.allowedAll) || ce.matchesField(ce.day, day)
	weekdayMatch := (len(ce.weekday.allowed) == 0 && !ce.weekday.allowedAll) || ce.matchesField(ce.weekday, weekday)

	if len(ce.day.allowed) == 0 && len(ce.weekday.allowed) == 0 {
		return true
	}
	if len(ce.day.allowed) > 0 && len(ce.weekday.allowed) > 0 {
		return dayMatch || weekdayMatch
	}
	if len(ce.day.allowed) > 0 {
		return dayMatch
	}
	return weekdayMatch
}
