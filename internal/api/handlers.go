package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"redisolar-go/internal/dao"
	redisdao "redisolar-go/internal/dao/redis"
	"redisolar-go/internal/models"
)

const (
	defaultRecentFeeds = 100
	maxRecentFeeds     = 1000
	defaultMetricCount = 120
	defaultCapLimit    = 10
	defaultRadius      = 10.0
	defaultGeoUnit     = "km"
)

func getFeedCount(count int) int {
	if count <= 0 {
		return defaultRecentFeeds
	}
	if count > maxRecentFeeds {
		return maxRecentFeeds
	}
	return count
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

// extractSiteID gets the site_id from the URL path like /sites/123 or /metrics/123
func extractIDFromPath(path string, prefix string) (int, bool) {
	if len(path) <= len(prefix) {
		return 0, false
	}
	idStr := path[len(prefix):]
	// Strip trailing slash if present
	if len(idStr) > 0 && idStr[len(idStr)-1] == '/' {
		idStr = idStr[:len(idStr)-1]
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, false
	}
	return id, true
}

// --- Site handlers ---

func siteListHandler(siteDao *redisdao.SiteDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sites, err := siteDao.FindAll(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sitesToResponse(sites))
	}
}

func siteByIDHandler(siteDao *redisdao.SiteDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := extractIDFromPath(r.URL.Path, "/sites/")
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid site id")
			return
		}
		site, err := siteDao.FindByID(r.Context(), id)
		if err == dao.ErrSiteNotFound {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, siteToResponse(site))
	}
}

// --- Site Geo handlers ---

func siteGeoListHandler(geoDao *redisdao.SiteGeoDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lngStr := r.URL.Query().Get("lng")

		if latStr == "" && lngStr == "" {
			sites, err := geoDao.FindAll(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, sitesToResponse(sites))
			return
		}

		if latStr == "" || lngStr == "" {
			writeError(w, http.StatusNotFound, "both lat and lng required")
			return
		}

		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid lat")
			return
		}
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid lng")
			return
		}

		radius := defaultRadius
		if r.URL.Query().Get("radius") != "" {
			radius, _ = strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
		}

		radiusUnit := defaultGeoUnit
		if r.URL.Query().Get("radius_unit") != "" {
			radiusUnit = r.URL.Query().Get("radius_unit")
		}

		onlyExcess := r.URL.Query().Get("only_excess_capacity") == "true"

		query := models.GeoQuery{
			Coordinate:         models.Coordinate{Lat: lat, Lng: lng},
			Radius:             radius,
			RadiusUnit:         models.GeoUnit(radiusUnit),
			OnlyExcessCapacity: onlyExcess,
		}

		sites, err := geoDao.FindByGeo(r.Context(), query)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sitesToResponse(sites))
	}
}

func siteGeoByIDHandler(geoDao *redisdao.SiteGeoDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := extractIDFromPath(r.URL.Path, "/sites/")
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid site id")
			return
		}
		site, err := geoDao.FindByID(r.Context(), id)
		if err == dao.ErrSiteNotFound {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, siteToResponse(site))
	}
}

// --- Capacity Report handler ---

func capacityReportHandler(capDao *redisdao.CapacityReportDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := defaultCapLimit
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		report, err := capDao.GetReport(r.Context(), limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, capacityReportToResponse(report))
	}
}

// --- Meter Reading handlers ---

func meterReadingPostHandler(meterReadingDao *redisdao.MeterReadingDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var envelope MeterReadingsEnvelope
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		for _, dto := range envelope.Readings {
			reading := dtoToMeterReading(dto)
			if err := meterReadingDao.Add(r.Context(), reading); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		writeJSON(w, http.StatusAccepted, "Accepted")
	}
}

func globalFeedHandler(feedDao *redisdao.FeedDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count := 0
		if c := r.URL.Query().Get("count"); c != "" {
			count, _ = strconv.Atoi(c)
		}
		readings, err := feedDao.GetRecentGlobal(r.Context(), getFeedCount(count))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, MeterReadingsEnvelope{Readings: meterReadingsToDTO(readings)})
	}
}

func siteFeedHandler(feedDao *redisdao.FeedDaoRedis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := extractIDFromPath(r.URL.Path, "/meter_readings/")
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid site id")
			return
		}
		count := 0
		if c := r.URL.Query().Get("count"); c != "" {
			count, _ = strconv.Atoi(c)
		}
		readings, err := feedDao.GetRecentForSite(r.Context(), id, getFeedCount(count))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, MeterReadingsEnvelope{Readings: meterReadingsToDTO(readings)})
	}
}

// --- Metrics handler ---

func metricsHandler(metricDao *redisdao.MetricDaoRedisTimeseries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := extractIDFromPath(r.URL.Path, "/metrics/")
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid site id")
			return
		}

		count := defaultMetricCount
		if c := r.URL.Query().Get("count"); c != "" {
			if parsed, err := strconv.Atoi(c); err == nil && parsed > 0 {
				count = parsed
			}
		}

		now := time.Now().UTC()
		generated, err := metricDao.GetRecent(r.Context(), id, models.WHGenerated, now, count)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		used, err := metricDao.GetRecent(r.Context(), id, models.WHUsed, now, count)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		plots := PlotsResponse{
			Plots: []PlotDTO{
				plotToDTO(models.Plot{Name: "Watt-Hours Generated", Measurements: generated}),
				plotToDTO(models.Plot{Name: "Watt-Hours Used", Measurements: used}),
			},
		}
		writeJSON(w, http.StatusOK, plots)
	}
}
