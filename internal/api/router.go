package api

import (
	"net/http"

	redisdao "redisolar-go/internal/dao/redis"
)

type Deps struct {
	SiteDao         *redisdao.SiteDaoRedis
	SiteGeoDao      *redisdao.SiteGeoDaoRedis
	CapacityDao     *redisdao.CapacityReportDaoRedis
	MetricDao       *redisdao.MetricDaoRedisTimeseries
	FeedDao         *redisdao.FeedDaoRedis
	MeterReadingDao *redisdao.MeterReadingDaoRedis
	UseGeoSiteAPI   bool
	StaticDir       string
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()

	// Static files: HTML references /static/css/..., /static/js/... etc.
	// Strip "/static/" prefix and serve from the StaticDir on disk.
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(deps.StaticDir)))
	mux.Handle("/static/", fs)

	// Sites routes - conditional on geo API
	if deps.UseGeoSiteAPI {
		mux.HandleFunc("/sites/", siteGeoByIDHandler(deps.SiteGeoDao))
		mux.HandleFunc("/sites", siteGeoListHandler(deps.SiteGeoDao))
	} else {
		mux.HandleFunc("/sites/", siteByIDHandler(deps.SiteDao))
		mux.HandleFunc("/sites", siteListHandler(deps.SiteDao))
	}

	// Capacity
	mux.HandleFunc("/capacity", capacityReportHandler(deps.CapacityDao))

	// Meter readings - need to distinguish between POST, GET, and GET with ID
	mux.HandleFunc("/meter_readings/", siteFeedHandler(deps.FeedDao))
	mux.HandleFunc("/meter_readings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			meterReadingPostHandler(deps.MeterReadingDao)(w, r)
		case http.MethodGet:
			globalFeedHandler(deps.FeedDao)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Metrics
	mux.HandleFunc("/metrics/", metricsHandler(deps.MetricDao))

	// Root serves index.html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, deps.StaticDir+"/index.html")
	})

	// Apply middleware
	var handler http.Handler = mux
	handler = loggingMiddleware(handler)
	handler = corsMiddleware(handler)

	return handler
}
