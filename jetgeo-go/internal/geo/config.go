package geo

import (
	"os"
	"time"
)

// Config 运行配置
// 优先级: 显式传入 > 环境变量 > 默认值

type Config struct {
	GeoDataPath string
	Level       Level
	CityTTL     time.Duration
	DistrictTTL time.Duration
}

func DefaultConfig() Config {
	return Config{
		GeoDataPath: "./data/geodata",
		Level:       LevelDistrict,
		CityTTL:     5 * time.Minute,
		DistrictTTL: 5 * time.Minute,
	}
}

func LoadFromEnv(base Config) Config {
	cfg := base
	if v := os.Getenv("GEO_DATA_PATH"); v != "" {
		cfg.GeoDataPath = v
	}
	if v := os.Getenv("JETGEO_LEVEL"); v != "" {
		if lv, err := ParseLevel(v); err == nil {
			cfg.Level = lv
		}
	}
	return cfg
}
