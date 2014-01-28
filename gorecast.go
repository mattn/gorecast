package main

import (
	"database/sql"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Metrics struct {
	Id          int64  `db:"id" json:"id"`
	ServiceName string `db:"service_name" json:"service_name"`
	SectionName string `db:"section_name" json:"section_name"`
	GraphName   string `db:"graph_name" json:"graph_name"`
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	imageDir := filepath.Join(cwd, "tmp")

	db, err := sql.Open("sqlite3", "gorecast.db")
	if err != nil {
		log.Fatal(err)
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	defer dbmap.Db.Close()

	dbmap.AddTableWithName(Metrics{}, "metrics").SetKeys(true, "id")
	//dbmap.AddTableWithName(Data{}, "data").SetKeys(true, "metrics_id")
	dbmap.AddTableWithName(Data{}, "data")
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		log.Fatal(err)
	}

	m := martini.Classic()
	m.Use(martini.Static("public"))
	m.Use(render.Renderer(render.Options{
		//Directory: "templates", // Specify what path to load the templates from.
		Layout: "layout", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
		//Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
		//Funcs: []template.FuncMap{AppHelpers}, // Specify helper function maps for templates to access.
		//Delims: render.Delims{"{[{", "}]}"}, // Sets delimiters to the specified strings.
		//Charset: "UTF-8", // Sets encoding for json and html content-types. Default is "UTF-8".
		//IndentJSON: true, // Output human readable JSON
	}))
	//m.MapTo(dbmap, (*gorp.DbMap)(nil))
	m.Get("/", func(renderer render.Render) {
		var metrics []Metrics
		_, err := dbmap.Select(&metrics, "select * from metrics")
		if err != nil {
			renderer.Error(400)
			return
		}
		renderer.HTML(200, "index", map[string]interface{}{
			"title":   "foo",
			"metrics": metrics,
		})
	})
	m.Get("/image/:service/:section/:graph.png", func(res http.ResponseWriter, params martini.Params) {
		res.Header().Set("Content-Type", "image/png")
		name := filepath.Join(imageDir, fmt.Sprintf("%s_%s_%s.png", params["service"], params["section"], params["graph"]))
		f, err := os.Open(name)
		if err != nil {
			http.Error(res, err.Error(), 400)
			return
		}
		defer f.Close()
		io.Copy(res, f)
	})
	m.Get("/api/:service", func(renderer render.Render, params martini.Params) {
		var metrics []Metrics
		_, err := dbmap.Select(&metrics, "select * from metrics where service_name = ?", params["service"])
		if err != nil {
			renderer.Error(400)
			return
		}
		renderer.JSON(200, metrics)
	})
	m.Get("/api/:service/:section", func(renderer render.Render, params martini.Params) {
		var metrics []Metrics
		_, err := dbmap.Select(&metrics, "select * from metrics where service_name = ? and section_name = ?", params["service"], params["section"])
		if err != nil {
			renderer.Error(400)
			return
		}
		renderer.JSON(200, metrics)
	})
	m.Get("/api/:service/:section/:graph", func(renderer render.Render, params martini.Params) {
		var metrics Metrics
		err := dbmap.SelectOne(&metrics, "select * from metrics where service_name = ? and section_name = ? and graph_name = ?", params["service"], params["section"], params["graph"])
		if err != nil {
			renderer.Error(400)
			return
		}
		renderer.JSON(200, metrics)
	})
	m.Post("/api/:service/:section/:graph", func(req *http.Request, renderer render.Render, params martini.Params) {
		if err != nil {
			log.Println(err.Error())
			renderer.Error(400)
			return
		}
		number := req.FormValue("number")
		f, err := strconv.ParseFloat(number, 64)
		if err != nil {
			log.Println(err.Error())
			renderer.Error(400)
			return
		}
		var metrics Metrics
		err = dbmap.SelectOne(&metrics, "select * from metrics where service_name = ? and section_name = ? and graph_name = ?", params["service"], params["section"], params["graph"])
		if err != nil {
			log.Println(err.Error())
			renderer.Error(400)
			return
		}

		var data Data
		data.MetricsId = metrics.Id
		data.DateTime = time.Now()
		data.UpdatedAt = time.Now()
		data.Number = f
		err = dbmap.Insert(&data)
		if err != nil {
			log.Println(err.Error())
			renderer.Error(400)
			return
		}
		renderer.JSON(200, data)
	})

	timer := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-timer.C
			var metrics []Metrics
			_, err := dbmap.Select(&metrics, "select * from metrics")
			if err != nil {
				continue
			}
			var wg sync.WaitGroup
			wg.Add(len(metrics))
			for _, metric := range metrics {
				go func(metric *Metrics) {
					graph(dbmap, metric, imageDir)
					wg.Done()
				}(&metric)
			}
			wg.Wait()
			log.Println("generated images")
		}
	}()

	m.Run()
}
