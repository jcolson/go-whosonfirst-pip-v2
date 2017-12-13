package index

// https://s2geometry.io/
// https://github.com/golang/geo
// https://godoc.org/github.com/golang/geo/s2

import (
	"errors"
	"github.com/golang/geo/s2"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	// "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"	
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

type GeoJSONShape struct {
	s2.Shape
	feature geojson.Feature
}

func LoopFromPolygon(p geom.Polygon) *s2.Loop {

	points := make([]s2.Point, 0)

	for _, coord := range p.Vertices() {

		ll := s2.LatLngFromDegrees(coord.Y, coord.X)
		pt := s2.PointFromLatLng(ll)

		points = append(points, pt)
	}

	return s2.LoopFromPoints(points)
}

func NewGeoJSONShapeForFeature(f geojson.Feature) (s2.Shape, error) {

	var sh s2.Shape
	var err error

	switch geometry.Type(f) {

	case "Polygon":
		sh, err = NewGeoJSONShapeForPolygonFeature(f)
	default:
		err = errors.New("Unsupport geometry type")
	}

	return sh, err
}

/*
func NewGeoJSONShapeForPointFeature(f geojson.Feature) (s2.Shape, error) {

	// what to do if not WOF???

	centroid, err := whosonfirst.Centroid(f)

	if err != nil {
		return nil, err
	}

	coord := centroid.Coord()

	ll := s2.LatLngFromDegrees(coord.Y, coord.X)
	pt := s2.PointFromLatLng(ll)

	cap := s2.CapFromPoint(pt)
	
	return cap, nil
}
*/

func NewGeoJSONShapeForPolygonFeature(f geojson.Feature) (s2.Shape, error) {

	polys, err := f.Polygons()

	if err != nil {
		return nil, err
	}

	loops := make([]*s2.Loop, 0)

	for _, p := range polys {

		ext_ring := p.ExteriorRing()
		ext_loop := LoopFromPolygon(ext_ring)

		if !ext_loop.IsValid() {
			return nil, errors.New("Invalid exterior ring")
		}

		loops = append(loops, ext_loop)

		for _, int_ring := range p.InteriorRings() {

			int_loop := LoopFromPolygon(int_ring)

			if !int_loop.IsValid() {
				return nil, errors.New("Invalid interior ring")
			}

			loops = append(loops, int_loop)
		}
	}

	sh := s2.PolygonFromLoops(loops)
	return sh, nil
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
