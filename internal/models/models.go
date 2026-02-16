package models

import "time"

// MetricUnit represents supported measurement metrics.
type MetricUnit string

const (
	WHGenerated MetricUnit = "whG"
	WHUsed      MetricUnit = "whU"
	TempCelsius MetricUnit = "tempC"
)

// GeoUnit represents geographic units available for geo queries.
type GeoUnit string

const (
	GeoUnitM  GeoUnit = "m"
	GeoUnitKM GeoUnit = "km"
	GeoUnitMI GeoUnit = "mi"
	GeoUnitFT GeoUnit = "ft"
)

// Coordinate represents a geographic coordinate pair.
type Coordinate struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// Site represents a solar power installation.
type Site struct {
	ID         int         `json:"id"`
	Capacity   float64     `json:"capacity"`
	Panels     int         `json:"panels"`
	Address    string      `json:"address"`
	City       string      `json:"city"`
	State      string      `json:"state"`
	PostalCode string      `json:"postal_code"`
	Coordinate *Coordinate `json:"coordinate,omitempty"`
}

// SiteCapacityTuple represents capacity at a site.
type SiteCapacityTuple struct {
	Capacity float64 `json:"capacity"`
	SiteID   int     `json:"site_id"`
}

// CapacityReport represents a site capacity report.
type CapacityReport struct {
	HighestCapacity []SiteCapacityTuple `json:"highest_capacity"`
	LowestCapacity  []SiteCapacityTuple `json:"lowest_capacity"`
}

// GeoQuery represents parameters for a geo query.
type GeoQuery struct {
	Coordinate         Coordinate
	Radius             float64
	RadiusUnit         GeoUnit
	OnlyExcessCapacity bool
}

// Measurement represents a measurement taken for a site.
type Measurement struct {
	SiteID     int        `json:"site_id"`
	Value      float64    `json:"value"`
	MetricUnit MetricUnit `json:"metric_unit"`
	Timestamp  float64    `json:"timestamp"`
}

// MeterReading represents a reading taken from a site.
type MeterReading struct {
	SiteID      int     `json:"site_id"`
	WHUsed      float64 `json:"wh_used"`
	WHGenerated float64 `json:"wh_generated"`
	TempC       float64 `json:"temp_c"`
	Timestamp   float64 `json:"timestamp"`
}

// CurrentCapacity returns the current capacity for this reading.
func (mr MeterReading) CurrentCapacity() float64 {
	return mr.WHGenerated - mr.WHUsed
}

// TimestampTime returns the timestamp as a time.Time.
func (mr MeterReading) TimestampTime() time.Time {
	sec := int64(mr.Timestamp)
	nsec := int64((mr.Timestamp - float64(sec)) * 1e9)
	return time.Unix(sec, nsec)
}

// Plot represents a plot of measurements.
type Plot struct {
	Measurements []Measurement `json:"measurements"`
	Name         string        `json:"name"`
}

// SiteStats field name constants matching Python.
const (
	SiteStatsLastReportingTime = "last_reporting_time"
	SiteStatsCount             = "meter_reading_count"
	SiteStatsMaxWH             = "max_wh_generated"
	SiteStatsMinWH             = "min_wh_generated"
	SiteStatsMaxCapacity       = "max_capacity"
)

// SiteStats represents reporting stats for a site.
type SiteStats struct {
	LastReportingTime string  `json:"last_reporting_time"`
	MeterReadingCount int64   `json:"meter_reading_count"`
	MaxWHGenerated    float64 `json:"max_wh_generated"`
	MinWHGenerated    float64 `json:"min_wh_generated"`
	MaxCapacity       float64 `json:"max_capacity"`
}

// Hash field name constants for Site (flat representation).
const (
	SiteFieldID         = "id"
	SiteFieldCapacity   = "capacity"
	SiteFieldPanels     = "panels"
	SiteFieldAddress    = "address"
	SiteFieldCity       = "city"
	SiteFieldState      = "state"
	SiteFieldPostalCode = "postal_code"
	SiteFieldLat        = "lat"
	SiteFieldLng        = "lng"
)
