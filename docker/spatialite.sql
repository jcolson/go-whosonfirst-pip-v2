SELECT load_extension('mod_spatialite');
CREATE TABLE geometries (
	id INTEGER NOT NULL PRIMARY KEY,
	is_alt TINYINT,
	type TEXT,
	lastmodified INTEGER
);

SELECT InitSpatialMetaData();
SELECT AddGeometryColumn('geometries', 'geom', 4326, 'GEOMETRY', 'XY');
SELECT CreateSpatialIndex('geometries', 'geom');

CREATE INDEX geometries_by_lastmod ON geometries (lastmodified);

CREATE VIRTUAL TABLE search USING fts4(
	id, placetype,
	name, names_all, names_preferred, names_variant, names_colloquial,		
	is_current, is_ceased, is_deprecated, is_superseded
);
