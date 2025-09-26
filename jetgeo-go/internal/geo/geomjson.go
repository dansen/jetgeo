package geo

import (
	"encoding/json"
	"io"
	"os"
)

type jsonPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func parsePolygonsFromFile(path string) ([][]jsonPoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parsePolygons(f)
}

func parsePolygons(r io.Reader) ([][]jsonPoint, error) {
	var polys [][]jsonPoint
	dec := json.NewDecoder(r)
	if err := dec.Decode(&polys); err != nil {
		return nil, err
	}
	return polys, nil
}
