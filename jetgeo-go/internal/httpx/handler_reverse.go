package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/dansen/jetgeo-go/internal/geo"
)

type ReverseHandler struct {
	Engine *geo.Engine
	Log    *zap.Logger
}

func NewRouter(h *ReverseHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(RequestLogger(h.Log))
	r.Get("/api/reverse", h.handleReverse)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return r
}

func (h *ReverseHandler) handleReverse(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	h.Log.Debug("reverse_request", zap.String("lat", latStr), zap.String("lng", lngStr))
	if latStr == "" || lngStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("missing lat or lng"))
		h.Log.Warn("reverse_missing_param")
		return
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid lat or lng"))
		h.Log.Warn("reverse_parse_error", zap.Error(err1), zap.Error(err2))
		return
	}
	gi, ok := h.Engine.Reverse(lat, lng)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		h.Log.Info("reverse_not_found", zap.Float64("lat", lat), zap.Float64("lng", lng))
		return
	}
	h.Log.Info("reverse_hit",
		zap.Float64("lat", lat),
		zap.Float64("lng", lng),
		zap.String("province", gi.Province),
		zap.String("provinceCode", gi.ProvinceCode),
		zap.String("city", gi.City),
		zap.String("cityCode", gi.CityCode),
		zap.String("district", gi.District),
		zap.String("districtCode", gi.DistrictCode),
		zap.String("adcode", gi.Adcode),
		zap.String("level", gi.Level.String()),
	)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(gi)
}
