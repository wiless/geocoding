package main

import (
	"fmt"
	"log"

	"golang.org/x/image/colornames"

	"github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"

	"github.com/fogleman/gg"
)

func main() {

	tp := sm.GetTileProviders()
	// log.Printf("Available Tile Providers %#v ", tp)
	// log.Println(tp["opentopomap"])

	// for k, v := range tp {
	k := "thunderforest-landscape"
	v := tp[k]
	ctx := sm.NewContext()
	// zoom := 1
	ctx.SetSize(600, 600)
	// pos := s2.LatLngFromDegrees(52.514536, 13.350151)

	mk := s2.LatLngFromDegrees(23.004345, 72.620645)

	ctx.SetCenter(s2.LatLngFromDegrees(23.005076, 72.621096))
	area, _ := sm.ParseAreaString("23.005076, 72.621096")
	ctx.AddArea(area)
	ctx.SetZoom(15)
	ctx.SetTileProvider(v)

	m := sm.NewMarker(mk, colornames.Red, 10)
	m.Label = "Home"
	ctx.AddMarker(m)
	ctx.AddPath(Paths())
	// ctx.AddMarker(sm.NewMarker(pos, color.RGBA{0xff, 0, 0, 0xff}, 16.0))
	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}

	fname := fmt.Sprintf("map-%s.png", k)
	log.Println("Check File ", fname)
	if err := gg.SavePNG(fname, img); err != nil {
		panic(err)
	}
	// }
}

func Paths() (path *sm.Path) {
	// pts := []string{"23.006270, 72.619419", "23.007040, 72.622112"}
	path = new(sm.Path)
	path.Weight = 2.0
	result := make([]s2.LatLng, 5)
	result[0] = s2.LatLngFromDegrees(23.006270, 72.619419)
	result[1] = s2.LatLngFromDegrees(23.006270, 72.622112)
	result[2] = s2.LatLngFromDegrees(23.007040, 72.622112)
	result[3] = s2.LatLngFromDegrees(23.007040, 72.619419)
	result[4] = result[0]
	path.Color = colornames.Green
	path.Positions = result
	return path
}
