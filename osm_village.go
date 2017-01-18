package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/image/colornames"

	"github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
	"github.com/wiless/vlib"

	"github.com/fogleman/gg"
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

var NEAR int = 2

type GPmap map[int]GPInfo
type VLmap map[int]VillageInfo

func main() {

	GPMAP := ReadGPTable("goaGP.csv")

	GPMAP.ProcessNeighbours()

	VLMAP := ReadVillageTable("goaVillage.csv")
	VLMAP.ProcessVillageDistances(GPMAP)

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

	ctx.SetCenter(s2.LatLngFromDegrees(15.461566, 73.9940975))
	// area, _ := sm.ParseAreaString("23.005076, 72.621096")

	// ctx.AddArea(area)
	// ctx.SetZoom(11)
	ctx.SetTileProvider(v)

	for k, v := range GPMAP {
		_ = k
		mk := sm.NewMarker(v.Location, colornames.Yellow, 10)
		// mk.Label = fmt.Sprintf("%s", v.Name)
		ctx.AddMarker(mk)
	}

	for k, v := range VLMAP {
		_ = k
		mk := sm.NewMarker(v.Location, colornames.Black, 5)
		// mk.Label = fmt.Sprintf("%s", v.Name)
		ctx.AddMarker(mk)
	}

	for k, v := range VLMAP {
		_ = k

		if v.ClosestGPdistance > 3.0 {
			path := new(sm.Path)
			path.Weight = 2.0
			result := make([]s2.LatLng, 2)

			result[0] = v.Location
			result[1] = GPMAP[v.ClosestGP].Location
			path.Color = colornames.Green
			path.Positions = result
			ctx.AddPath(path)
		}

		if v.AdminDist > 3.0 && v.AdminGP != v.ClosestGP {
			path := new(sm.Path)
			path.Weight = 2.0
			path.Color = colornames.Red
			result := make([]s2.LatLng, 2)

			result[0] = v.Location
			result[1] = GPMAP[v.AdminGP].Location
			path.Positions = result
			ctx.AddPath(path)
		}

		mk := sm.NewMarker(v.Location, colornames.Black, 5)
		// mk.Label = fmt.Sprintf("%s", v.Name)
		ctx.AddMarker(mk)
	}

	// ctx.AddMarker(sm.NewMarker(pos, color.RGBA{0xff, 0, 0, 0xff}, 16.0))
	img, err := ctx.Render()

	if err != nil {
		panic(err)
	}

	fname := fmt.Sprintf("GoaVillages-%s.png", k)
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

type Location struct {
	Lat float64
	Lng float64
}

func Distance(p1, p2 s2.LatLng) float64 {

	a, b := math.Pi/2.0-p1.Lat.Radians(), math.Pi/2.0-p2.Lat.Radians()
	c := (p1.Lng - p2.Lng).Radians()
	input := math.Cos(a)*math.Cos(b) + math.Sin(a)*math.Sin(b)*math.Cos(c)

	if input > 1 {
		input = 1
	}
	if input < -1 {
		input = -1
	}
	d := math.Acos(input) * 6371.0
	if math.IsNaN(d) {
		log.Println(d, input)
	}

	return d
}

type VillageInfo struct {
	VillageID         int
	IsGP              bool
	Name              string
	AdminGP           int
	AdminDist         float64
	ClosestGP         int
	ClosestGPdistance float64
	distance          vlib.VectorF
	Neighbours        vlib.VectorI
	Location          s2.LatLng
}

func (vmap VLmap) ProcessVillageDistances(gmap GPmap) {

	for k, v := range vmap {
		gpinfo, ok := gmap[v.AdminGP]
		if ok {

			v.AdminDist = Distance(v.Location, gpinfo.Location)
			radioranges := vlib.NewVectorF(len(gmap))
			gpids := vlib.NewVectorI(len(gmap))
			cnt := 0

			for i, g := range gmap {
				radioranges[cnt] = Distance(v.Location, g.Location)
				gpids[cnt] = i

				cnt++
			}

			sdist := vlib.NewVSliceF(radioranges...)
			sort.Sort(sdist)
			sindx := sdist.SIndex()
			v.distance = radioranges[0:NEAR]
			v.Neighbours = gpids.At(sindx[0:NEAR]...)

			// log.Printf("Min distance %d with GP %d : %v", k, gpids[sindx[0]], radioranges.Min())
			v.ClosestGP = v.Neighbours[0]
			v.ClosestGPdistance = v.distance[0]
			vmap[k] = v
		} else {
			log.Println("Administrive GP info not found ", k, v.AdminGP)
		}
	}

}

func (gpm GPmap) ProcessNeighbours() {
	for k, v := range gpm {
		v.Neighbours = make([]int, len(gpm)-1)

		v.distance.Resize(len(gpm) - 1)
		v.distance.Fill(99999)

		cnt := 0
		invalidData := true
		for i, u := range gpm {
			if i == k {

				continue
			} else {
				v.Neighbours[cnt] = i
				v.distance[cnt] = Distance(v.Location, u.Location)

				if v.distance[cnt] == 0 {
					fmt.Println("REALLY !! ", v.Location, u.Location, i, k)
					invalidData = false

				}

			}
			cnt++

		}

		v.IsValid = invalidData
		s := vlib.NewVSliceF(v.distance...)
		sort.Sort(s)
		sindex := s.SIndex()
		// log.Println(v.Neighbours.At(sindex[0:10]...))
		v.distance = v.distance[0:10]
		v.Neighbours = v.Neighbours.At(sindex[0:10]...)
		v.ClosestGPdistance = v.distance[0]
		v.ClosestGP = v.Neighbours[0]

		// fmt.Println("=========  ============ == = = ====")
		v.ClosestGP = v.Neighbours[0]
		// pretty.Println(v)
		// time.Sleep(1 * time.Second)
		gpm[k] = v

	}

}

func ReadVillageTable(fname string) (result VLmap) {

	fd, ferr := os.Open(fname)
	defer fd.Close()
	if ferr != nil {
		log.Panicln("Error Opening File ", ferr)
	}
	rd := csv.NewReader(fd)
	rd.TrailingComma = true
	rd.TrimLeadingSpace = true

	result = make(map[int]VillageInfo)

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
		var vlinfo VillageInfo
		vlinfo.VillageID, _ = strconv.Atoi(strings.TrimSpace(record[0]))
		vlinfo.Name = strings.TrimSpace(record[1])
		vlinfo.AdminGP, _ = strconv.Atoi(strings.TrimSpace(record[2]))
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(record[4]), 64)
		vlinfo.Location = s2.LatLngFromDegrees(lat, lng)
		if (err1 != nil) || (err2 != nil) {
			log.Println("======= ========  Error", record, lat, lng, err1, err2)
		} else {

			_, ok := result[vlinfo.VillageID]
			if ok {
				log.Println("Duplicate Village Entry Found ", vlinfo.VillageID, ok)
				dcnt++
			}
			result[vlinfo.VillageID] = vlinfo
			// fmt.Println(gpinfo)
			cnt++

		}
		total++
		// fmt.Println("Distance ", Distance(gpinfo.Location, reflat))

	}
	log.Printf("Villages Processed %d of %d records [%d Duplicates] in %s ", cnt, total, dcnt, fname)

	return result
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
