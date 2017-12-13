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
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

func LoopFromPolygon(p geom.Polygon) *s2.Loop {

	points := make([]s2.Point, 0)

	for _, coord := range p.Vertices() {

		ll := s2.LatLngFromDegrees(coord.Y, coord.X)
		pt := s2.PointFromLatLng(ll)

		points = append(points, pt)
	}

	return s2.LoopFromPoints(points)
}

func ShapesForFeature(f geojson.Feature) ([]s2.Shape, error) {

	var sh []s2.Shape
	var err error

	switch geometry.Type(f) {

	case "Polygon":
		sh, err = ShapesForPolygonFeature(f)
	case "MultiPolygon":
		sh, err = ShapesForPolygonFeature(f)
	case "Point":
		err = errors.New("Unsupported geometry type")
	default:
		err = errors.New("Unsupported geometry type")
	}

	return sh, err
}

func ShapesForPolygonFeature(f geojson.Feature) ([]s2.Shape, error) {

	polys, err := f.Polygons()

	if err != nil {
		return nil, err
	}

	shapes := make([]s2.Shape, 0)

	for _, p := range polys {

		loops := make([]*s2.Loop, 0)

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

		sh := s2.PolygonFromLoops(loops)
		shapes = append(shapes, sh)
	}

	return shapes, nil
}

type S2Index struct {
	Index
	shapeindex *s2.ShapeIndex
	cache      cache.Cache
}

func NewS2Index(c cache.Cache) (Index, error) {

	si := s2.NewShapeIndex()

	i := S2Index{
		shapeindex: si,
		cache:      c,
	}

	return &i, nil
}

func (i *S2Index) IndexFeature(f geojson.Feature) error {

	shapes, err := ShapesForFeature(f)

	if err != nil {
		return err
	}

	// question: where do I assign/append the WOF ID for the polygon/shape
	// being added to the index? (20171212/thisisaaronland)

	for _, sh := range shapes {
		i.shapeindex.Add(sh)
	}

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
