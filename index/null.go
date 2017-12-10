package index

// this is basically just here to make setting up new indexes easier
// (20171210/thisisaaronland)

import (
	"errors"
	"github.com/skelterjohn/geom"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-pip"
	"github.com/whosonfirst/go-whosonfirst-pip/cache"
	"github.com/whosonfirst/go-whosonfirst-pip/filter"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

type NullIndex struct {
	Index
}

func (i *NullIndex) IndexFeature(f geojson.Feature) error {
	return errors.New("Please write me")
}

func (i *NullIndex) Cache() cache.Cache {
	return nil
}

func (i *NullIndex) GetIntersectsByCoord(geom.Coord, filter.Filter) (spr.StandardPlacesResults, error) {
	return nil, errors.New("Please write me")
}

func (i *NullIndex) GetCandidatesByCoord(geom.Coord) (*pip.GeoJSONFeatureCollection, error) {
	return nil, errors.New("Please write me")
}

func (i *NullIndex) GetIntersectsByPath(geom.Path, filter.Filter) ([]spr.StandardPlacesResults, error) {
	return nil, errors.New("Please write me")
}
