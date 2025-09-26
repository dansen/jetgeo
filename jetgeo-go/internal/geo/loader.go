package geo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/geo/s2"
)

type regionMeta struct {
	Level      Level
	ParentCode string
	Code       string
	Name       string
}

func parseFileName(name string) (regionMeta, error) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.Split(base, "_")
	if len(parts) < 3 {
		return regionMeta{}, fmt.Errorf("invalid filename: %s", name)
	}
	switch parts[0] {
	case "province":
		if len(parts) < 3 {
			return regionMeta{}, fmt.Errorf("invalid province filename: %s", name)
		}
		return regionMeta{Level: LevelProvince, Code: parts[1], Name: parts[2]}, nil
	case "city":
		if len(parts) < 4 {
			return regionMeta{}, fmt.Errorf("invalid city filename: %s", name)
		}
		return regionMeta{Level: LevelCity, ParentCode: parts[1], Code: parts[2], Name: parts[3]}, nil
	case "district":
		if len(parts) < 4 {
			return regionMeta{}, fmt.Errorf("invalid district filename: %s", name)
		}
		return regionMeta{Level: LevelDistrict, ParentCode: parts[1], Code: parts[2], Name: parts[3]}, nil
	default:
		return regionMeta{}, fmt.Errorf("unknown level in filename: %s", name)
	}
}

// loadSingleRegion 构造 RegionCache
func loadSingleRegion(path string) (*RegionCache, error) {
	polys, err := parsePolygonsFromFile(path)
	if err != nil {
		return nil, err
	}
	meta, err := parseFileName(filepath.Base(path))
	if err != nil {
		return nil, err
	}
	var polygons []*s2.Polygon
	for _, ring := range polys { // 每个 ring 一个 polygon (与 Java 单环一致)
		pts := make([]s2.Point, 0, len(ring))
		for _, p := range ring {
			ll := s2.LatLngFromDegrees(p.Lat, p.Lng).Normalized()
			pts = append(pts, ll.Point())
		}
		loop := s2.LoopFromPoints(pts)
		if loop.IsHole() {
			loop.Invert()
		}
		poly := s2.PolygonFromLoops([]*s2.Loop{loop})
		polygons = append(polygons, poly)
	}
	return &RegionCache{Name: meta.Name, ParentCode: meta.ParentCode, Code: meta.Code, Level: meta.Level, Polygons: polygons}, nil
}

// LoadProvinces 预加载全部省级
func LoadProvinces(root string) (map[string]*RegionCache, error) {
	dir := filepath.Join(root, "province")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*RegionCache)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		rc, err := loadSingleRegion(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		res[rc.Code] = rc
	}
	return res, nil
}

// LoadChildren 加载 city/district 子级
func LoadChildren(root string, level Level, parentCode string) ([]*RegionCache, error) {
	dir := filepath.Join(root, level.String())
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var list []*RegionCache
	prefix := level.String() + "_" + parentCode + "_"
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		if !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		rc, err := loadSingleRegion(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		list = append(list, rc)
	}
	return list, nil
}
