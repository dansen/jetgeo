package geo

import "github.com/golang/geo/s2"

// GeoInfo 反向地理结果
// 与 Java 版本字段保持一致的 json tag
// Level 用字符串序列化由 handler 手动转换
type GeoInfo struct {
	FormatAddress string `json:"formatAddress"`
	Province      string `json:"province"`
	ProvinceCode  string `json:"provinceCode"`
	City          string `json:"city"`
	CityCode      string `json:"cityCode"`
	District      string `json:"district"`
	DistrictCode  string `json:"districtCode"`
	Street        string `json:"street"`
	StreetCode    string `json:"streetCode"`
	Adcode        string `json:"adcode"`
	Level         Level  `json:"level"`
}

// RegionCache 单个行政区域的多边形缓存
// Polygons: 一个区域可能多个不连续多边形
type RegionCache struct {
	Name       string
	ParentCode string
	Code       string
	Level      Level
	Polygons   []*s2.Polygon
}

func (r *RegionCache) CacheKey() string {
	if r.ParentCode != "" {
		return r.ParentCode + ":" + r.Code
	}
	return r.Code
}

// Contains 判断点是否在任一 Polygon 内
func (r *RegionCache) Contains(lat, lng float64) bool {
	if len(r.Polygons) == 0 {
		return false
	}
	p := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng).Normalized())
	for _, poly := range r.Polygons {
		if poly.ContainsPoint(p) {
			return true
		}
	}
	return false
}
