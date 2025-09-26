package geo

import "fmt"

// Engine 对应 Java JetGeo

type Engine struct {
	cfg           Config
	province      map[string]*RegionCache
	cityCache     *LoadingCache
	districtCache *LoadingCache
}

func NewEngine(cfg Config) (*Engine, error) {
	prov, err := LoadProvinces(cfg.GeoDataPath)
	if err != nil {
		return nil, fmt.Errorf("load provinces: %w", err)
	}
	eng := &Engine{cfg: cfg, province: prov}
	if cfg.Level >= LevelCity {
		eng.cityCache = NewLoadingCache(cfg.CityTTL, func(parent string) ([]*RegionCache, error) {
			return LoadChildren(cfg.GeoDataPath, LevelCity, parent)
		})
	}
	if cfg.Level >= LevelDistrict {
		eng.districtCache = NewLoadingCache(cfg.DistrictTTL, func(parent string) ([]*RegionCache, error) {
			return LoadChildren(cfg.GeoDataPath, LevelDistrict, parent)
		})
	}
	return eng, nil
}

// Reverse 获取行政信息
func (e *Engine) Reverse(lat, lng float64) (*GeoInfo, bool) {
	var province *RegionCache
	for _, rc := range e.province {
		if rc.Contains(lat, lng) {
			province = rc
			break
		}
	}
	if province == nil {
		return nil, false
	}
	gi := &GeoInfo{Province: province.Name, ProvinceCode: province.Code, Adcode: province.Code, Level: LevelProvince}
	// city
	if e.cfg.Level >= LevelCity && e.cityCache != nil {
		if cities, err := e.cityCache.Get(province.Code); err == nil {
			for _, c := range cities {
				if c.Contains(lat, lng) {
					gi.City = c.Name
					gi.CityCode = c.Code
					gi.Adcode = c.Code
					gi.Level = LevelCity
					break
				}
			}
		}
	}
	// district
	if gi.Adcode != "" && e.cfg.Level >= LevelDistrict && e.districtCache != nil {
		if districts, err := e.districtCache.Get(gi.Adcode); err == nil {
			for _, d := range districts {
				if d.Contains(lat, lng) {
					gi.District = d.Name
					gi.DistrictCode = d.Code
					gi.Adcode = d.Code
					gi.Level = LevelDistrict
					break
				}
			}
		}
	}
	gi.FormatAddress = gi.Province + gi.City + gi.District
	return gi, true
}
