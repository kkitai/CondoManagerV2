package domain

import "time"

type PropertyStatus string

const (
	PropertyStatusActive   PropertyStatus = "active"
	PropertyStatusInactive PropertyStatus = "inactive"
)

type Property struct {
	ID                int64
	Name              string
	Address           string
	Area              *float64
	UnitCount         *int
	Status            PropertyStatus
	ManagementCompany *string
	AssigneeID        *int64
	Assignee          *User
	CreatedBy         *int64
	UpdatedBy         *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PropertyStats struct {
	TotalClaims        int64
	OpenClaims         int64
	InProgressClaims   int64
	CompletedClaims    int64
	CategoryBreakdown  []CategoryCount
	SeverityBreakdown  []SeverityCount
	AvgResponseDays    *float64
	RecurrenceRate     *float64
}

type CategoryCount struct {
	Category string
	Count    int64
}

type SeverityCount struct {
	Severity string
	Count    int64
}

type PropertyListStats struct {
	TotalCount    int64
	ActiveCount   int64
	InactiveCount int64
	TotalUnits    int64
	AvgUnits      float64
}

type PropertyListParams struct {
	Search          string
	Status          string
	AssigneeID      int64
	Area            string
	UnitCountMin    int
	UnitCountMax    int
	UpdatedFrom     time.Time
	UpdatedTo       time.Time
	Page            int
	PerPage         int
	SortColumn      string
	SortOrder       string
}
