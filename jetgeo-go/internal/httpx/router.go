package httpx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dansen/jetgeo-go/internal/geo"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GinHandler struct {
	Engine *geo.Engine
	Log    *zap.Logger
}

// NewGinEngine builds gin.Engine with middleware and routes
func NewGinEngine(h *GinHandler) *gin.Engine {
	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(requestLogger(h.Log))
	g.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	g.GET("/api/reverse", h.reverse)
	return g
}

// requestLogger custom zap logging middleware (similar to previous one)
func requestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		log.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("duration", dur),
			zap.String("remote", c.ClientIP()),
		)
	}
}

func (h *GinHandler) reverse(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	h.Log.Debug("reverse_request", zap.String("lat", latStr), zap.String("lng", lngStr))
	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, Err[any](1001, "missing lat or lng"))
		h.Log.Warn("reverse_missing_param")
		return
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, Err[any](1002, "invalid lat or lng"))
		h.Log.Warn("reverse_parse_error", zap.Error(err1), zap.Error(err2))
		return
	}
	gi, ok := h.Engine.Reverse(lat, lng)
	if !ok {
		c.JSON(http.StatusNotFound, Err[any](1404, "not found"))
		h.Log.Info("reverse_not_found", zap.Float64("lat", lat), zap.Float64("lng", lng))
		return
	}
	// wrap
	c.JSON(http.StatusOK, Ok(gi))
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
}
