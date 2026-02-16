package models

import (
	"fmt"
	"strconv"
)

// SiteToFlatMap converts a Site to a flat map for Redis HASH storage.
func SiteToFlatMap(s Site) map[string]interface{} {
	m := map[string]interface{}{
		SiteFieldID:         s.ID,
		SiteFieldCapacity:   s.Capacity,
		SiteFieldPanels:     s.Panels,
		SiteFieldAddress:    s.Address,
		SiteFieldCity:       s.City,
		SiteFieldState:      s.State,
		SiteFieldPostalCode: s.PostalCode,
	}
	if s.Coordinate != nil {
		m[SiteFieldLat] = s.Coordinate.Lat
		m[SiteFieldLng] = s.Coordinate.Lng
	}
	return m
}

// SiteFromFlatMap converts a Redis HASH (flat map) back to a Site.
func SiteFromFlatMap(m map[string]string) (Site, error) {
	id, err := strconv.Atoi(m[SiteFieldID])
	if err != nil {
		return Site{}, fmt.Errorf("invalid site id: %w", err)
	}
	capacity, err := strconv.ParseFloat(m[SiteFieldCapacity], 64)
	if err != nil {
		return Site{}, fmt.Errorf("invalid capacity: %w", err)
	}
	panels, err := strconv.Atoi(m[SiteFieldPanels])
	if err != nil {
		return Site{}, fmt.Errorf("invalid panels: %w", err)
	}

	site := Site{
		ID:         id,
		Capacity:   capacity,
		Panels:     panels,
		Address:    m[SiteFieldAddress],
		City:       m[SiteFieldCity],
		State:      m[SiteFieldState],
		PostalCode: m[SiteFieldPostalCode],
	}

	latStr, hasLat := m[SiteFieldLat]
	lngStr, hasLng := m[SiteFieldLng]
	if hasLat && hasLng && latStr != "" && lngStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			return Site{}, fmt.Errorf("invalid lat: %w", err)
		}
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			return Site{}, fmt.Errorf("invalid lng: %w", err)
		}
		site.Coordinate = &Coordinate{Lat: lat, Lng: lng}
	}

	return site, nil
}

// MeterReadingToStreamMap converts a MeterReading to a map for Redis STREAM.
func MeterReadingToStreamMap(mr MeterReading) map[string]interface{} {
	return map[string]interface{}{
		"site_id":      mr.SiteID,
		"wh_used":      mr.WHUsed,
		"wh_generated": mr.WHGenerated,
		"temp_c":       mr.TempC,
		"timestamp":    mr.Timestamp,
	}
}

// MeterReadingFromStreamMap converts a Redis STREAM map back to a MeterReading.
func MeterReadingFromStreamMap(m map[string]interface{}) (MeterReading, error) {
	siteID, err := toInt(m["site_id"])
	if err != nil {
		return MeterReading{}, fmt.Errorf("invalid site_id: %w", err)
	}
	whUsed, err := toFloat64(m["wh_used"])
	if err != nil {
		return MeterReading{}, fmt.Errorf("invalid wh_used: %w", err)
	}
	whGenerated, err := toFloat64(m["wh_generated"])
	if err != nil {
		return MeterReading{}, fmt.Errorf("invalid wh_generated: %w", err)
	}
	tempC, err := toFloat64(m["temp_c"])
	if err != nil {
		return MeterReading{}, fmt.Errorf("invalid temp_c: %w", err)
	}
	timestamp, err := toFloat64(m["timestamp"])
	if err != nil {
		return MeterReading{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	return MeterReading{
		SiteID:      siteID,
		WHUsed:      whUsed,
		WHGenerated: whGenerated,
		TempC:       tempC,
		Timestamp:   timestamp,
	}, nil
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	case int64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func toInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}
