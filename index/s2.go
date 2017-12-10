package index

// https://s2geometry.io/
// https://github.com/golang/geo

import (
	"errors"
	"github.com/golang/geo/s2"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

type GeoJSONShape struct {
	s2.Shape
}

/*
func (sh *GeoJSONShape) NumEdges() int {
     return 0			// please write m
}

func (sh *GeoJSONShape) Edge(i int) s2.Edge {
     return nil		       // please write me
}

func (sh *GeoJSONShape) HasInterior() bool {
     return false		// please write me
}

func (sh *GeoJSONShape) ReferencePoint() s2.ReferencePoint {
     return nil			// please write me
}

func (sh *GeoJSONShape) NumChains() int {
     return 0			// please write me
}

func (sh *GeoJSONShape) Chain(chainID int) s2.Chain {
     return nil			// please write me
}

func (sh *GeoJSONShape) ChainEdge(chainID, offset int) s2.Edge {
     return nil			// please write me
}

func (sh *GeoJSONShape) ChainPosition(edgeID int) ChainPosition {
     return nil			// please write me
}

*/

func NewGeoJSONShapeForFeature(f geojson.Feature) (s2.Shape, error) {

	return nil, errors.New("Please write me")
}

type S2Index struct {
	Index
	shapeindex *s2.ShapeIndex
}

func NewS2Index() (Index, error) {

	si := s2.NewShapeIndex()

	i := S2Index{
		shapeindex: si,
	}

	return &i, nil
}

func (i *S2Index) IndexFeature(f geojson.Feature) error {

	sh, err := NewGeoJSONShapeForFeature(f)

	if err != nil {
		return err
	}

	i.shapeindex.Add(sh)
	return nil
}

func (i *S2Index) Cache() cache.Cache {
	return nil
}

func (i *S2Index) GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error) {
	return nil, errors.New("Please write me")
}

func (i *S2Index) GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error) {
	return nil, errors.New("Please write me")
}

func (i *S2Index) GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error) {
	return nil, errors.New("Please write me")
}
