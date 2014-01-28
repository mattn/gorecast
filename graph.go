package main

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	//"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var font *truetype.Font

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	b, err := ioutil.ReadFile(filepath.Join(cwd, "fonts", "ipaexg.ttf"))
	if err != nil {
		log.Fatal(err)
	}
	font, err = freetype.ParseFont(b)
	if err != nil {
		log.Println(err)
	}
}

type Data struct {
	MetricsId int64     `db:"metrics_id" json:"metrics_id"`
	DateTime  time.Time `db:"datetime" json:"datetime"`
	Number    float64   `db:"number" json:"number"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func graph(dbmap *gorp.DbMap, metrics *Metrics, outdir string) error {
	datas := []Data{}
	_, err := dbmap.Select(&datas, "select * from data where metrics_id = ? order by datetime", metrics.Id)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	rgba := image.NewRGBA(image.Rect(0, 0, 400, 300))
	draw.Draw(rgba, rgba.Bounds(), image.White, image.ZP, draw.Src)
	img := imgg.AddTo(rgba, 0, 0, 400, 280, color.RGBA{0xff, 0xff, 0xff, 0xff}, font, imgg.ConstructFontSizes(13))

	dt := make([]chart.EPoint, 0, 20)
	for _, data := range datas {
		dt = append(dt, chart.EPoint{
			X: float64(data.DateTime.Unix()),
			Y: float64(data.Number),
		})
	}

	c := chart.ScatterChart{Title: metrics.GraphName}
	c.XRange.TicSetting.Grid = 1
	if len(dt) > 0 {
		c.AddData("", dt, chart.PlotStyleLinesPoints, chart.Style{})
	}
	c.XRange.Time = true
	c.XRange.TicSetting.TFormat = func(t time.Time, td chart.TimeDelta) string {
		return t.Format("15:04")
	}
	c.YRange.Label = metrics.GraphName
	c.Plot(img)
	f, err := os.Create(filepath.Join(outdir, fmt.Sprintf("%s_%s_%s.png", metrics.ServiceName, metrics.SectionName, metrics.GraphName)))
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, rgba)
}
