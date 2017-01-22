package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/colornames"

	"github.com/bmatsuo/csvutil"
	coordsparser "github.com/flopp/go-coordsparser"
	"github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"github.com/wiless/vlib"
)

type GPInfo struct {
	GPID              int
	Name              string
	ClosestGP         int
	ClosestGPdistance float64
	distance          vlib.VectorF
	Neighbours        vlib.VectorI
	Location          s2.LatLng
	IsValid           bool
}

type GPmap map[int]GPInfo

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
	// gps, pos := ReadAllGP()
	if len(os.Args) < 2 {
		log.Println("Pass input file name (e.g ./fetchmap Google162_Lucknow.csv) ")
		return
	}
	fname := os.Args[1]

	gpmap := ReadGPTable(fname)

	//ctx.SetCenter(s2.LatLngFromDegrees(27.018090, 80.918009))
	// p1 := s2.LatLngFromDegrees(27.018090, 80.918009)
	// p2 := s2.LatLngFromDegrees(27.134025, 80.906947)

	/// TRIAL FOR BOUNDARY
	var dr s2.Rect
	var nulllat s2.LatLng
	// dr := s2.RectFromLatLng(p1)
	fmt.Println("Start =========")
	for _, v := range gpmap {

		/// load boundary
		if dr.Lo() == nulllat {
			dr = s2.RectFromLatLng(v.Location)
		} else {
			dr = dr.AddPoint(v.Location)
		}

		// Load Markers
		mk := sm.NewMarker(v.Location, colornames.Yellow, 10)
		// mk.Label = fmt.Sprintf("%s", v.Name)
		ctx.AddMarker(mk)

	}
	boundary := dr.RectBound()
	ctx.SetCenter(boundary.Center())
	fmt.Println("FULL BOUNDARY", dr.RectBound())
	path := new(sm.Path)
	{

		path.Weight = 1.0
		result := make([]s2.LatLng, 5)
		for i := 0; i < 4; i++ {
			result[i] = boundary.Vertex(i)
		}
		result[4] = result[0]
		path.Color = colornames.Green
		path.Positions = result

	}
	// pp1 := s2.PointFromLatLng(p1)
	// pp2 := s2.PointFromLatLng(p2)
	// fmt.Printf("Point 1 =%v \n Point 2 = %v ", pp1, pp2)

	// area, _ := sm.ParseAreaString("23.005076, 72.621096")

	// ctx.AddArea(area)
	// ctx.SetZoom(11)

	// CREATE MAP

	ctx.SetTileProvider(v)

	// for k, v := range gpmap {
	// 	_ = k
	// 	mk := sm.NewMarker(v.Location, colornames.Yellow, 10)
	// 	// mk.Label = fmt.Sprintf("%s", v.Name)
	// 	ctx.AddMarker(mk)
	// }

	ctx.AddPath(path)
	// // ctx.AddMarker(sm.NewMarker(pos, color.RGBA{0xff, 0, 0, 0xff}, 16.0))
	img, err := ctx.Render()

	if err != nil {
		panic(err)
	}

	finfo, _ := os.Stat(fname)
	mfname := fmt.Sprintf("Map-%s.png", finfo.Name())
	log.Println("Check File ", mfname)
	if err := gg.SavePNG(mfname, img); err != nil {
		panic(err)
	}

}

func CreateBoundary() {

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

func ReadGPTable(fname string) (result GPmap) {

	// reflat := s2.LatLngFromDegrees(16.8070444, 75.2322867)
	fd, ferr := os.Open(fname)
	defer fd.Close()
	if ferr != nil {
		log.Panicln("Error Opening File ", ferr)
	}
	rd := csv.NewReader(fd)
	rd.TrailingComma = true
	rd.TrimLeadingSpace = true

	result = make(map[int]GPInfo)

	// skipping the header
	rd.Read()
	cnt := 0
	total := 0
	dcnt := 0
	// log.Println("Header ", rowstr)
	for {
		record, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var gpinfo GPInfo
		gpinfo.IsValid = true
		gpinfo.GPID, _ = strconv.Atoi(record[0])
		gpinfo.Name = strings.TrimSpace(record[1])
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(record[4]), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(record[5]), 64)
		gpinfo.Location = s2.LatLngFromDegrees(lat, lng)
		if (err1 != nil) || (err2 != nil) {
			log.Println("** Error Processing Location ", record, lat, lng, err1, err2)
		} else {

			_, ok := result[gpinfo.GPID]
			if ok {
				log.Println("Duplicate GP Entry Found ", gpinfo.GPID, ok)
				dcnt++
			}
			result[gpinfo.GPID] = gpinfo
			// fmt.Println(gpinfo)
			cnt++

		}
		total++
		// fmt.Println("Distance ", Distance(gpinfo.Location, reflat))

	}
	log.Printf("GPs Processed %d of %d records [%d dups] in %s ", cnt, total, dcnt, fname)

	return result
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
