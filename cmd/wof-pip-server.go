package main

import (
	"fmt"
	"github.com/whosonfirst/go-http-mapzenjs"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pip/app"
	"github.com/whosonfirst/go-whosonfirst-pip/flags"
	"github.com/whosonfirst/go-whosonfirst-pip/http"
	"io"
	golog "log"
	gohttp "net/http"
	"os"
	"runtime"
	godebug "runtime/debug"
	"time"
)

func main() {

	fl, err := flags.CommonFlags()

	if err != nil {
		golog.Fatal(err)
	}

	fl.String("host", "localhost", "The hostname to listen for requests on")
	fl.Int("port", 8080, "The port number to listen for requests on")

	fl.Bool("enable-geojson", false, "Allow users to request GeoJSON FeatureCollection formatted responses.")
	fl.Bool("enable-extras", false, "")
	fl.Bool("enable-candidates", false, "")
	fl.Bool("enable-polylines", false, "")
	fl.Bool("enable-www", false, "")

	fl.Int("polylines-coords", 100, "...")
	fl.String("www-path", "/debug", "...")

	flags.Parse(fl, os.Args[1:])

	verbose, _ := flags.BoolVar(fl, "verbose")
	procs, _ := flags.IntVar(fl, "processes")

	logger := log.SimpleWOFLogger()
	level := "status"

	if verbose {
		level = "debug"
	}

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, level)

	runtime.GOMAXPROCS(procs)

	enable_www, _ := flags.BoolVar(fl, "enable_www")

	if enable_www {
		logger.Status("-enable-www flag is true causing the following flags to also be true: -enable-geojson -enable-candidates")

		fl.Set("enable_geojson", "true")
		fl.Set("enable_candidates", "true")
	}

	pip_index, _ := flags.StringVar(fl, "index")
	pip_cache, _ := flags.StringVar(fl, "cache")
	mode, _ := flags.StringVar(fl, "mode")

	logger.Info("index is %s cache is %s mode is %s", pip_index, pip_cache, mode)

	logger.Debug("setting up application cache")

	appcache, err := app.NewApplicationCache(fl)

	if err != nil {
		logger.Fatal("Failed to create caching layer, because %s", err)
	}

	logger.Debug("setting up application index")

	appindex, err := app.NewApplicationIndex(fl, appcache)

	if err != nil {
		logger.Fatal("Failed to create indexing layer, because %s", err)
	}

	indexer, err := app.NewApplicationIndexer(fl, appindex)

	if err != nil {
		logger.Fatal("Failed to create indexer, because %s", err)
	}

	// note: this is "-mode spatialite" not "-engine spatialite"

	if mode != "spatialite" {

		go func() {

			// TO DO: put this somewhere so that it can be triggered by signal(s)
			// to reindex everything in bulk or incrementally

			t1 := time.Now()

			err = indexer.IndexPaths(fl.Args())

			if err != nil {
				logger.Fatal("failed to index paths because %s", err)
			}

			t2 := time.Since(t1)

			logger.Status("finished indexing in %v", t2)
			godebug.FreeOSMemory()
		}()

		// set up some basic monitoring and feedback stuff

		go func() {

			c := time.Tick(1 * time.Second)

			for _ = range c {

				if !indexer.IsIndexing() {
					continue
				}

				logger.Status("indexing %d records indexed", indexer.Indexed)
			}
		}()
	}

	go func() {

		tick := time.Tick(1 * time.Minute)

		for _ = range tick {
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			logger.Status("memstats system: %8d inuse: %8d released: %8d objects: %6d", ms.HeapSys, ms.HeapInuse, ms.HeapReleased, ms.HeapObjects)
		}
	}()

	// set up the HTTP endpoint

	logger.Debug("setting up intersects handler")

	enable_geojson, _ := flags.BoolVar(fl, "enable-geojson")
	enable_extras, _ := flags.BoolVar(fl, "enable_extras")
	extras_dsn, _ := flags.StringVar(fl, "extras_dsn")

	// enable_extras is set above...

	intersects_opts := http.NewDefaultIntersectsHandlerOptions()
	intersects_opts.EnableGeoJSON = enable_geojson
	intersects_opts.EnableExtras = enable_extras
	intersects_opts.ExtrasDB = extras_dsn

	intersects_handler, err := http.IntersectsHandler(appindex, indexer, intersects_opts)

	if err != nil {
		logger.Fatal("failed to create PIP handler because %s", err)
	}

	ping_handler, err := http.PingHandler()

	if err != nil {
		logger.Fatal("failed to create Ping handler because %s", err)
	}

	mux := gohttp.NewServeMux()

	mux.Handle("/ping", ping_handler)
	mux.Handle("/", intersects_handler)

	enable_candidates, _ := flags.BoolVar(fl, "enable-candidates")
	enable_polylines, _ := flags.BoolVar(fl, "enable-polylines")

	// enable_www is set above

	if enable_candidates {

		logger.Debug("setting up candidates handler")

		candidateshandler, err := http.CandidatesHandler(appindex, indexer)

		if err != nil {
			logger.Fatal("failed to create Spatial handler because %s", err)
		}

		mux.Handle("/candidates", candidateshandler)
	}

	if enable_polylines {

		logger.Debug("setting up polylines handler")

		poly_coords, _ := flags.IntVar(fl, "polylines-coords")

		poly_opts := http.NewDefaultPolylineHandlerOptions()
		poly_opts.MaxCoords = poly_coords
		poly_opts.EnableGeoJSON = enable_geojson

		poly_handler, err := http.PolylineHandler(appindex, indexer, poly_opts)

		if err != nil {
			logger.Fatal("failed to create polyline handler because %s", err)
		}

		mux.Handle("/polyline", poly_handler)
	}

	if enable_www {

		logger.Debug("setting up www handler")

		var www_handler gohttp.Handler

		bundled_handler, err := http.BundledWWWHandler()

		if err != nil {
			logger.Fatal("failed to create (bundled) www handler because %s", err)
		}

		www_handler = bundled_handler

		/*

			mapzenjs_opts := mapzenjs.DefaultMapzenJSOptions()
			mapzenjs_opts.APIKey = *www_apikey

			mapzenjs_handler, err := mapzenjs.MapzenJSHandler(www_handler, mapzenjs_opts)

			if err != nil {
				logger.Fatal("failed to create mapzen.js handler because %s", err)
			}

				mzjs_opts := mapzenjs.DefaultMapzenJSOptions()
				mzjs_opts.APIKey = *api_key

				mzjs_handler, err := mapzenjs.MapzenJSHandler(www_handler, mzjs_opts)

				if err != nil {
					logger.Fatal("failed to create API key handler because %s", err)
				}

				opts := rewrite.DefaultRewriteRuleOptions()

				rewrite_path := *www_path

				if strings.HasSuffix(rewrite_path, "/") {
					rewrite_path = strings.TrimRight(rewrite_path, "/")
				}

				rule := rewrite.RemovePrefixRewriteRule(rewrite_path, opts)
				rules := []rewrite.RewriteRule{rule}

				debug_handler, err := rewrite.RewriteHandler(rules, apikey_handler)

				if err != nil {
					logger.Fatal("failed to create www handler because %s", err)
				}
		*/

		mapzenjs_assets_handler, err := mapzenjs.MapzenJSAssetsHandler()

		if err != nil {
			logger.Fatal("failed to create mapzenjs_assets handler because %s", err)
		}

		mux.Handle("/javascript/mapzen.min.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/tangram.min.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/mapzen.js", mapzenjs_assets_handler)
		mux.Handle("/javascript/tangram.js", mapzenjs_assets_handler)
		mux.Handle("/css/mapzen.js.css", mapzenjs_assets_handler)
		mux.Handle("/tangram/refill-style.zip", mapzenjs_assets_handler)

		mux.Handle("/javascript/mapzen.whosonfirst.pip.js", www_handler)
		mux.Handle("/javascript/slippymap.crosshairs.js", www_handler)
		mux.Handle("/css/mapzen.whosonfirst.pip.css", www_handler)

		www_path, _ := flags.StringVar(fl, "www-path")
		mux.Handle(www_path, www_handler)
	}

	host, _ := flags.StringVar(fl, "host")
	port, _ := flags.IntVar(fl, "port")

	endpoint := fmt.Sprintf("%s:%d", host, port)
	logger.Status("listening for requests on %s", endpoint)

	err = gohttp.ListenAndServe(endpoint, mux)

	if err != nil {
		logger.Fatal("failed to start server because %s", err)
	}

	os.Exit(0)
}
