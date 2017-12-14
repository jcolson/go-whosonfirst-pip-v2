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
	golog "log"
)

func PointFromCoord(coord geom.Coord) s2.Point {

	ll := s2.LatLngFromDegrees(coord.Y, coord.X)
	pt := s2.PointFromLatLng(ll)

	return pt
}

func LoopFromPolygon(p geom.Polygon) *s2.Loop {

	vertices := p.Vertices()
	count := len(vertices)

	points := make([]s2.Point, 0)

	for i, coord := range vertices {

		pt := PointFromCoord(coord)

		// see notes in ShapesForPolygons() below

		if (i + 1) != count {
			points = append(points, pt)
		}
	}

	return s2.LoopFromPoints(points)
}

func ShapesForFeature(f geojson.Feature) ([]s2.Shape, error) {

	var sh []s2.Shape
	var err error

	t := geometry.Type(f)
	// log.Println("TYPE", t)

	switch t {

	case "Polygon":
		sh, err = ShapesForPolygonFeature(f)
	case "MultiPolygon":
		sh, err = ShapesForMultiPolygonFeature(f)
	case "Point":
		err = errors.New("Unsupported geometry type")
	default:
		err = errors.New("Unsupported geometry type")
	}

	return sh, err
}

// this is a separate function mostly just for clarity now and might just
// be merged with ShapesForPolygonFeature - the salient point is that the
// geojson.Polygons() method returns a flat list of geojson.Polygon thingies
// each of which has ExteriorRing() and InteriorRings() methods
// (20171214/thisisaaronland)

func ShapesForMultiPolygonFeature(f geojson.Feature) ([]s2.Shape, error) {

	polys, err := f.Polygons()

	if err != nil {
		return nil, err
	}

	return ShapesForPolygons(polys)
}

func ShapesForPolygonFeature(f geojson.Feature) ([]s2.Shape, error) {

	polys, err := f.Polygons()

	if err != nil {
		return nil, err
	}

	return ShapesForPolygons(polys)
}

/*

A linear ring MUST follow the right-hand rule with respect to the
area it bounds, i.e., exterior rings are counterclockwise, and
holes are clockwise.

Note: the [GJ2008] specification did not discuss linear ring winding
order.  For backwards compatibility, parsers SHOULD NOT reject
Polygons that do not follow the right-hand rule.

-- https://tools.ietf.org/html/rfc7946#section-3.1.6

An S2Loop represents a simple spherical polygon. It consists of a single
chain of vertices where the first vertex is implicitly connected to the last.
All loops are defined to have a CCW orientation, i.e. the interior of the loop
is on the left side of the edges. This implies that a clockwise loop enclosing
a small area is interpreted to be a CCW loop enclosing a very large area.

-- https://s2geometry.io/devguide/basic_types

Order of polygon vertices in general GIS: clockwise or counterclockwise

-- https://gis.stackexchange.com/questions/119150/order-of-polygon-vertices-in-general-gis-clockwise-or-counterclockwise

The polygon has a CW winding if the quantity in 1 is positive and a CCW winding otherwise.

-- http://blog.element84.com/polygon-winding.html

*/

func ShapesForPolygons(polys []geojson.Polygon) ([]s2.Shape, error) {

	shapes := make([]s2.Shape, 0)

	for _, p := range polys {

		loops := make([]*s2.Loop, 0)

		ext_ring := p.ExteriorRing()
		ext_loop := LoopFromPolygon(ext_ring)

		// ext_order := ext_ring.WindingOrder()

		// wuh... ?!
		// https://godoc.org/github.com/golang/geo/s2#Loop.IsValid

		if !ext_loop.IsValid() {
			return nil, errors.New("Invalid exterior ring")
		}

		loops = append(loops, ext_loop)

		for _, int_ring := range p.InteriorRings() {

			// int_order := ext_ring.WindingOrder()

			int_loop := LoopFromPolygon(int_ring)

			// wuh... ?!

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

		// we probably want to return an error here but while we're figuring out
		// how it all works we are just being chatty about it...
		// (20171214/thisisaaronland)
		// return err

		golog.Printf("SKIP %s (%s) because %s\n", f.Id(), f.Name(), err)
		return nil
	}

	// question: where do I assign/append the WOF ID for the polygon/shape
	// being added to the index? (20171212/thisisaaronland)

	for _, sh := range shapes {
		i.shapeindex.Add(sh)
	}

	str_id := f.Id()

	fc, err := cache.NewFeatureCache(f)

	if err != nil {
		return err
	}

	err = i.cache.Set(str_id, fc)

	if err != nil {
		return err
	}

	return nil
}

func (i *S2Index) Cache() cache.Cache {
	return i.cache
}

/*

I have no idea if I am doing the right thing here or why this isn't working - like
I am indexing things incorrectly (above) or do I want to be calling something other
than LocatePoint() or even using a ShapeIndexIterator at all? Either way assuming
we've indexed California correctly we should get a match for 37.794906,-122.395229

https://godoc.org/github.com/golang/geo/s2#ShapeIndexIterator

https://whosonfirst.mapzen.com/spelunker/id/420561633	<-- super bowl city
https://whosonfirst.mapzen.com/spelunker/id/85688637	<-- california

./bin/wof-pip -lru-cache -mode sqlite -latitude 37.794906 -longitude -122.395229 ~/Downloads/region-20171212.db
2017/12/14 09:22:03 SKIP 85667653 (Malanje) because Invalid interior ring
2017/12/14 09:22:05 SKIP 85668081 (Ciudad de Buenos Aires) because Invalid interior ring
2017/12/14 09:22:05 SKIP 85668095 (Tavush) because Invalid interior ring
2017/12/14 09:22:05 SKIP 85668215 (Goranboy) because Invalid interior ring
...time passes and California is not in this list...
2017/12/14 09:24:26 SKIP 85688607 (New Jersey) because Invalid interior ring
2017/12/14 09:24:31 SKIP 85688805 (Kiev City) because Invalid interior ring
2017/12/14 09:24:32 SKIP 85688825 (Chernihiv) because Invalid interior ring
2017/12/14 09:24:33 SKIP 85688877 (Biliaivskyi) because Invalid interior ring
2017/12/14 09:24:39 BY COORD {-122.395229 37.794906} (-0.423359865649069688764428, -0.667231204056549009884236, 0.612836800862064379202820) false

(20171214/thisisaaronland)

*/

func (i *S2Index) GetIntersectsByCoord(c geom.Coord, f filter.Filter) (spr.StandardPlacesResults, error) {

	pt := PointFromCoord(c)

	iter := s2.NewShapeIndexIterator(i.shapeindex)
	ok := iter.LocatePoint(pt)

	golog.Println("BY COORD", c, pt, ok)

	return nil, errors.New("Please write me")
}

func (i *S2Index) GetCandidatesByCoord(c geom.Coord) (*pip.GeoJSONFeatureCollection, error) {
	return nil, errors.New("Please write me")
}

func (i *S2Index) GetIntersectsByPath(p geom.Path, f filter.Filter) ([]spr.StandardPlacesResults, error) {
	return nil, errors.New("Please write me")
}
