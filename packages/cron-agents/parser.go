package cron

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// ParsedSchedule represents a parsed cron expression.
type ParsedSchedule struct {
	// Original is the original cron expression string.
	Original string

	// Type indicates whether this is a standard or alias expression.
	Type string // "standard" or "alias"

	// Fields holds the parsed cron fields (minute, hour, day, month, weekday).
	Fields []string

	// Schedule is the underlying cron.Schedule implementation.
	Schedule cron.Schedule
}

// Parser handles parsing of cron expressions.
type Parser struct {
	// WithSeconds enables parsing of 6-field cron (with seconds).
	WithSeconds bool

	// parser is the underlying robfig/cron parser.
	parser cron.Parser
}

// NewParser creates a new cron parser.
func NewParser() *Parser {
	return &Parser{
		WithSeconds: false,
		parser:      cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// NewParserWithSeconds creates a parser that supports seconds.
func NewParserWithSeconds() *Parser {
	return &Parser{
		WithSeconds: true,
		parser:      cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// aliases maps human-readable expressions to cron syntax.
var aliases = map[string]string{
	"@yearly":    "0 0 1 1 *",
	"@annually":  "0 0 1 1 *",
	"@monthly":   "0 0 1 * *",
	"@weekly":    "0 0 * * 0",
	"@daily":     "0 0 * * *",
	"@midnight":  "0 0 * * *",
	"@hourly":    "0 * * * *",
	"@everymin":  "* * * * *",
	"@everyhour": "0 * * * *",
}

// predefinedTimezones maps common timezone aliases.
var predefinedTimezones = map[string]string{
	"UTC":  "UTC",
	"GMT":  "GMT",
	"EST":  "America/New_York",
	"PST":  "America/Los_Angeles",
	"PDT":  "America/Los_Angeles",
	"CST":  "America/Chicago",
	"IST":  "Asia/Kolkata",
	"JST":  "Asia/Tokyo",
	"CET":  "Europe/Paris",
	"BST":  "Europe/London",
	"AEST": "Australia/Sydney",
}

// Parse parses a cron expression (standard or alias).
func (p *Parser) Parse(expr string) (*ParsedSchedule, error) {
	if expr == "" {
		return nil, errors.New("empty cron expression")
	}

	expr = strings.TrimSpace(expr)
	original := expr

	// Check for @every syntax FIRST: @every <duration>
	if strings.HasPrefix(original, "@every ") {
		durationStr := strings.TrimPrefix(original, "@every ")
		return p.parseEvery(durationStr)
	}

	// Check for alias
	if strings.HasPrefix(expr, "@") {
		expanded, ok := aliases[strings.ToLower(expr)]
		if !ok {
			return nil, fmt.Errorf("unknown cron alias: %s", expr)
		}
		expr = expanded
	}

	// Parse using robfig/cron
	schedule, err := p.parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression '%s': %w", original, err)
	}

	// Determine expression type
	exprType := "standard"
	if strings.HasPrefix(original, "@") {
		exprType = "alias"
	}

	// Extract fields for inspection
	fields := strings.Fields(expr)

	return &ParsedSchedule{
		Original: original,
		Type:     exprType,
		Fields:   fields,
		Schedule: schedule,
	}, nil
}

// parseEvery handles @every <duration> syntax.
func (p *Parser) parseEvery(durationStr string) (*ParsedSchedule, error) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration '%s': %w", durationStr, err)
	}

	// Create a fixed interval schedule
	schedule := cron.ConstantDelaySchedule{Delay: duration}

	return &ParsedSchedule{
		Original: "@every " + durationStr,
		Type:     "interval",
		Fields:   []string{durationStr},
		Schedule: schedule,
	}, nil
}

// Next calculates the next run time after the given time.
func (ps *ParsedSchedule) Next(after time.Time) time.Time {
	if ps.Schedule == nil {
		return time.Time{}
	}
	return ps.Schedule.Next(after)
}

// NextInLocation calculates the next run time in the specified timezone.
func (ps *ParsedSchedule) NextInLocation(after time.Time, loc *time.Location) time.Time {
	if ps.Schedule == nil {
		return time.Time{}
	}
	// Convert to the target location for calculation
	afterInLoc := after.In(loc)
	return ps.Schedule.Next(afterInLoc)
}

// Describe returns a human-readable description of the schedule.
func (ps *ParsedSchedule) Describe() string {
	if ps.Type == "alias" {
		return describeAlias(ps.Original)
	}

	if ps.Type == "interval" {
		return ps.Original
	}

	return describeCronFields(ps.Fields)
}

// describeAlias returns a description for alias expressions.
func describeAlias(alias string) string {
	descriptions := map[string]string{
		"@yearly":   "once a year (January 1 at midnight)",
		"@annually": "once a year (January 1 at midnight)",
		"@monthly":  "once a month (1st at midnight)",
		"@weekly":   "once a week (Sunday at midnight)",
		"@daily":    "every day at midnight",
		"@midnight": "every day at midnight",
		"@hourly":   "every hour",
		"@everymin": "every minute",
	}

	if desc, ok := descriptions[strings.ToLower(alias)]; ok {
		return desc
	}
	return alias
}

// describeCronFields creates a human-readable description from cron fields.
func describeCronFields(fields []string) string {
	if len(fields) < 5 {
		return "invalid cron expression"
	}

	minute, hour, day, month, weekday := fields[0], fields[1], fields[2], fields[3], fields[4]

	var parts []string

	// Describe minute
	if minute == "*" {
		parts = append(parts, "every minute")
	} else if !isAllNumeric(minute) {
		parts = append(parts, fmt.Sprintf("at minute %s", minute))
	} else {
		m, _ := strconv.Atoi(minute)
		parts = append(parts, fmt.Sprintf("at minute %d", m))
	}

	// Describe hour
	if hour != "*" {
		h, _ := strconv.Atoi(hour)
		if hour != "*" && minute != "*" {
			parts = append(parts, fmt.Sprintf("past hour %d", h))
		}
	}

	// Describe day of month
	if day != "*" {
		parts = append(parts, fmt.Sprintf("on day %s of the month", day))
	}

	// Describe month
	if month != "*" {
		monthNames := []string{"", "January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"}
		if m, err := strconv.Atoi(month); err == nil && m >= 1 && m <= 12 {
			parts = append(parts, fmt.Sprintf("in %s", monthNames[m]))
		} else {
			parts = append(parts, fmt.Sprintf("in month %s", month))
		}
	}

	// Describe weekday
	if weekday != "*" {
		weekdayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		if w, err := strconv.Atoi(weekday); err == nil && w >= 0 && w <= 6 {
			parts = append(parts, fmt.Sprintf("on %s", weekdayNames[w]))
		} else {
			parts = append(parts, fmt.Sprintf("on weekday %s", weekday))
		}
	}

	return strings.Join(parts, " ")
}

// isAllNumeric checks if a string contains only digits.
func isAllNumeric(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}

// ValidateCronExpression validates a cron expression without parsing it.
func ValidateCronExpression(expr string) error {
	parser := NewParser()
	_, err := parser.Parse(expr)
	return err
}

// ExpandAlias expands a cron alias to its standard form.
func ExpandAlias(alias string) (string, bool) {
	if expanded, ok := aliases[strings.ToLower(alias)]; ok {
		return expanded, true
	}
	return "", false
}

// IsAlias checks if the expression is a known alias.
func IsAlias(expr string) bool {
	_, ok := aliases[strings.ToLower(expr)]
	return ok
}

// GetPredefinedTimezone returns the IANA timezone for a common abbreviation.
func GetPredefinedTimezone(abbr string) (string, bool) {
	if tz, ok := predefinedTimezones[strings.ToUpper(abbr)]; ok {
		return tz, true
	}
	return "", false
}

// ParseTime parses a time string in HH:MM format and returns hour and minute.
func ParseTime(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, errors.New("time must be in HH:MM format")
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	return hour, minute, nil
}

// CreateDailyCron creates a cron expression for daily execution at a specific time.
func CreateDailyCron(hour, minute int) string {
	return fmt.Sprintf("%d %d * * *", minute, hour)
}

// CreateWeeklyCron creates a cron expression for weekly execution.
// weekday: 0=Sunday, 1=Monday, ..., 6=Saturday
func CreateWeeklyCron(weekday, hour, minute int) string {
	return fmt.Sprintf("%d %d * * %d", minute, hour, weekday)
}

// CreateMonthlyCron creates a cron expression for monthly execution.
func CreateMonthlyCron(day, hour, minute int) string {
	return fmt.Sprintf("%d %d %d * *", minute, hour, day)
}

// ScheduleBuilder provides a fluent interface for building cron expressions.
type ScheduleBuilder struct {
	minute  string
	hour    string
	day     string
	month   string
	weekday string
}

// NewScheduleBuilder creates a new schedule builder.
func NewScheduleBuilder() *ScheduleBuilder {
	return &ScheduleBuilder{
		minute:  "*",
		hour:    "*",
		day:     "*",
		month:   "*",
		weekday: "*",
	}
}

// Minute sets the minute field.
func (b *ScheduleBuilder) Minute(minute int) *ScheduleBuilder {
	b.minute = strconv.Itoa(minute)
	return b
}

// Hour sets the hour field.
func (b *ScheduleBuilder) Hour(hour int) *ScheduleBuilder {
	b.hour = strconv.Itoa(hour)
	return b
}

// DayOfMonth sets the day of month field.
func (b *ScheduleBuilder) DayOfMonth(day int) *ScheduleBuilder {
	b.day = strconv.Itoa(day)
	return b
}

// Month sets the month field.
func (b *ScheduleBuilder) Month(month int) *ScheduleBuilder {
	b.month = strconv.Itoa(month)
	return b
}

// Weekday sets the day of week field (0=Sunday).
func (b *ScheduleBuilder) Weekday(weekday int) *ScheduleBuilder {
	b.weekday = strconv.Itoa(weekday)
	return b
}

// EveryMinute sets the schedule to run every minute.
func (b *ScheduleBuilder) EveryMinute() *ScheduleBuilder {
	b.minute = "*"
	return b
}

// EveryHour sets the schedule to run every hour.
func (b *ScheduleBuilder) EveryHour() *ScheduleBuilder {
	b.minute = "0"
	b.hour = "*"
	return b
}

// DailyAt sets the schedule to run daily at a specific time.
func (b *ScheduleBuilder) DailyAt(hour, minute int) *ScheduleBuilder {
	b.minute = strconv.Itoa(minute)
	b.hour = strconv.Itoa(hour)
	b.day = "*"
	b.month = "*"
	b.weekday = "*"
	return b
}

// WeeklyOn sets the schedule to run weekly on a specific day.
func (b *ScheduleBuilder) WeeklyOn(weekday, hour, minute int) *ScheduleBuilder {
	b.minute = strconv.Itoa(minute)
	b.hour = strconv.Itoa(hour)
	b.day = "*"
	b.month = "*"
	b.weekday = strconv.Itoa(weekday)
	return b
}

// MonthlyOn sets the schedule to run monthly on a specific day.
func (b *ScheduleBuilder) MonthlyOn(day, hour, minute int) *ScheduleBuilder {
	b.minute = strconv.Itoa(minute)
	b.hour = strconv.Itoa(hour)
	b.day = strconv.Itoa(day)
	b.month = "*"
	b.weekday = "*"
	return b
}

// Build creates the cron expression string.
func (b *ScheduleBuilder) Build() string {
	return fmt.Sprintf("%s %s %s %s %s", b.minute, b.hour, b.day, b.month, b.weekday)
}
