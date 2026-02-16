package api

import "redisolar-go/internal/models"

// SiteResponse is the JSON response for a single site (nested coordinate).
type SiteResponse struct {
	ID         int                `json:"id"`
	Capacity   float64            `json:"capacity"`
	Panels     int                `json:"panels"`
	Address    string             `json:"address"`
	City       string             `json:"city"`
	State      string             `json:"state"`
	PostalCode string             `json:"postal_code"`
	Coordinate *CoordinateDTO     `json:"coordinate,omitempty"`
}

type CoordinateDTO struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func siteToResponse(s models.Site) SiteResponse {
	r := SiteResponse{
		ID:         s.ID,
		Capacity:   s.Capacity,
		Panels:     s.Panels,
		Address:    s.Address,
		City:       s.City,
		State:      s.State,
		PostalCode: s.PostalCode,
	}
	if s.Coordinate != nil {
		r.Coordinate = &CoordinateDTO{
			Lng: s.Coordinate.Lng,
			Lat: s.Coordinate.Lat,
		}
	}
	return r
}

func sitesToResponse(sites []models.Site) []SiteResponse {
	result := make([]SiteResponse, len(sites))
	for i, s := range sites {
		result[i] = siteToResponse(s)
	}
	return result
}

// CapacityTupleDTO is the JSON representation of a capacity tuple.
type CapacityTupleDTO struct {
	Capacity float64 `json:"capacity"`
	SiteID   int     `json:"site_id"`
}

type CapacityReportResponse struct {
	HighestCapacity []CapacityTupleDTO `json:"highest_capacity"`
	LowestCapacity  []CapacityTupleDTO `json:"lowest_capacity"`
}

func capacityReportToResponse(r models.CapacityReport) CapacityReportResponse {
	highest := make([]CapacityTupleDTO, len(r.HighestCapacity))
	for i, t := range r.HighestCapacity {
		highest[i] = CapacityTupleDTO{Capacity: t.Capacity, SiteID: t.SiteID}
	}
	lowest := make([]CapacityTupleDTO, len(r.LowestCapacity))
	for i, t := range r.LowestCapacity {
		lowest[i] = CapacityTupleDTO{Capacity: t.Capacity, SiteID: t.SiteID}
	}
	return CapacityReportResponse{HighestCapacity: highest, LowestCapacity: lowest}
}

// MeterReadingsEnvelope wraps meter readings for JSON.
type MeterReadingsEnvelope struct {
	Readings []MeterReadingDTO `json:"readings"`
}

type MeterReadingDTO struct {
	SiteID      int     `json:"site_id"`
	WHUsed      float64 `json:"wh_used"`
	WHGenerated float64 `json:"wh_generated"`
	TempC       float64 `json:"temp_c"`
	Timestamp   float64 `json:"timestamp"`
}

func meterReadingToDTO(mr models.MeterReading) MeterReadingDTO {
	return MeterReadingDTO{
		SiteID:      mr.SiteID,
		WHUsed:      mr.WHUsed,
		WHGenerated: mr.WHGenerated,
		TempC:       mr.TempC,
		Timestamp:   mr.Timestamp,
	}
}

func meterReadingsToDTO(readings []models.MeterReading) []MeterReadingDTO {
	result := make([]MeterReadingDTO, len(readings))
	for i, r := range readings {
		result[i] = meterReadingToDTO(r)
	}
	return result
}

func dtoToMeterReading(dto MeterReadingDTO) models.MeterReading {
	return models.MeterReading{
		SiteID:      dto.SiteID,
		WHUsed:      dto.WHUsed,
		WHGenerated: dto.WHGenerated,
		TempC:       dto.TempC,
		Timestamp:   dto.Timestamp,
	}
}

// MeasurementDTO is the JSON representation of a measurement.
type MeasurementDTO struct {
	SiteID     int     `json:"site_id"`
	Value      float64 `json:"value"`
	MetricUnit string  `json:"metric_unit"`
	Timestamp  float64 `json:"timestamp"`
}

type PlotDTO struct {
	Measurements []MeasurementDTO `json:"measurements"`
	Name         string           `json:"name"`
}

type PlotsResponse struct {
	Plots []PlotDTO `json:"plots"`
}

func measurementToDTO(m models.Measurement) MeasurementDTO {
	return MeasurementDTO{
		SiteID:     m.SiteID,
		Value:      m.Value,
		MetricUnit: string(m.MetricUnit),
		Timestamp:  m.Timestamp,
	}
}

func plotToDTO(p models.Plot) PlotDTO {
	ms := make([]MeasurementDTO, len(p.Measurements))
	for i, m := range p.Measurements {
		ms[i] = measurementToDTO(m)
	}
	return PlotDTO{Measurements: ms, Name: p.Name}
}
