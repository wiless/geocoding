package main

import (
	"fmt"
	"log"

	"golang.org/x/image/colornames"

	"github.com/bmatsuo/csvutil"
	coordsparser "github.com/flopp/go-coordsparser"
	"github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"

	"github.com/fogleman/gg"
)

func main() {
	GetImage()
}

func GetImage() {

	tp := sm.GetTileProviders()
	// log.Printf("Available Tile Providers %#v ", tp)
	// log.Println(tp["opentopomap"])

	// for k, v := range tp {
	k := "thunderforest-landscape"
	v := tp[k]
	ctx := sm.NewContext()
	// zoom := 1
	ctx.SetSize(1024, 1024)
	// pos := s2.LatLngFromDegrees(52.514536, 13.350151)

	//mk := s2.LatLngFromDegrees(23.004345, 72.620645)
	// m := sm.NewMarker(mk, colornames.Red, 10)
	// m.Label = "Home"
	gps, pos := ReadAllGP()

	ctx.SetCenter(s2.LatLngFromDegrees(16.0742321, 74.7819819))
	// area, _ := sm.ParseAreaString("23.005076, 72.621096")
	area := new(sm.Area)
	area.Positions = pos
	area.Color = colornames.Yellow
	// ctx.AddArea(area)
	ctx.SetZoom(10)
	area.Fill = colornames.Aliceblue
	ctx.SetTileProvider(v)

	for _, gpm := range gps {
		ctx.AddMarker(gpm)

	}

	// ctx.AddMarker(sm.NewMarker(pos, color.RGBA{0xff, 0, 0, 0xff}, 16.0))
	img, err := ctx.Render()

	if err != nil {
		panic(err)
	}

	fname := fmt.Sprintf("GP-%s.png", k)
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

func ReadAllGP() (result []*sm.Marker, locations []s2.LatLng) {

	// r, ferr := os.Open("GP.csv")
	// log.Println(ferr)
	records, cerr := csvutil.ReadFile("GP.csv")
	if cerr != nil {
		log.Println("CSV err", cerr)
	}
	// csvutil.ReadAll("GP.csv")
	// var result []*sm.Marker
	locations = make([]s2.LatLng, 0)
	for i, r := range records {

		if i == 0 {
			continue
		}

		// log.Println(i, len(r), r[4], r[5])
		// lat, long := strconv.ParseFloat(r[4], 64)
		// log.Println("joined ", strings.Join(r[4:], ","))
		lat, lng, err := coordsparser.Parse(r[4] + "," + r[5])
		if err == nil {
			// fmt.Println(lat, lng)
			// fmt.Println("Entry ", i, r[4])
			// lat, lng, err := coordsparser.Parse(r[4])
			// if err != nil {
			// 	log.Println("Some error @ ", r[4], r[0], r[1], err)
			// }
			pos := s2.LatLngFromDegrees(lat, lng)
			mk := sm.NewMarker(pos, colornames.Red, 10)
			mk.Label = r[0]
			result = append(result, mk)
			locations = append(locations, pos)
		} else {
			log.Println("Skipping ", i, r[0:3], err)
		}

	}
	return result, locations
}
