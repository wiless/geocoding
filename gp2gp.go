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

	"github.com/golang/geo/s2"
	"github.com/wiless/vlib"
)

var NEAR int = 10

func main() {

	// reflat := s2.LatLngFromDegrees(15.0455814, 73.9888797)
	// reflat2 := s2.LatLngFromDegrees(15.0455814, 73.9888797)
	// fmt.Println(Distance(reflat, reflat), "IS NAN ? ")
	// fmt.Println(Distance(reflat, reflat2), "IS NAN ? ")

	GPMAP := ReadGPTable("goaGP.csv")

	GPMAP.ProcessNeighbours()

	VLMAP := ReadVillageTable("goaVillage.csv")
	VLMAP.ProcessVillageDistances(GPMAP)

	{
		// Dump GP to GP stats
		wd, _ := os.Create("goaGP2GP.csv")
		fmt.Fprintf(wd, "GPID, ClosestGP,ClosestGPdist, NGP1, NG2 , NGDist1, NGDist2")
		for k, g := range GPMAP {
			if g.IsValid {
				fmt.Fprintf(wd, "\n%d,%d,%f,%s,%s", k, g.ClosestGP, g.ClosestGPdistance, g.Neighbours.ToCSVStr(), g.distance.ToCSVStr())
			}
		}
		fmt.Println()

	}
	{
		// Dump Village Stats
		wd, _ := os.Create("goaStatistics.csv")
		fmt.Fprintf(wd, "VillageID, AdminGP ,AdminDistance, ClosestGP,ClosestGPdist, NGP1, NG2 , NGDist1, NGDist2")
		for k, v := range VLMAP {
			fmt.Fprintf(wd, "\n%d,%d,%f,%d,%f,%s,%s", k, v.AdminGP, v.AdminDist, v.ClosestGP, v.ClosestGPdistance, v.Neighbours.ToCSVStr(), v.distance.ToCSVStr())
		}
		fmt.Println()
	}

	str := `importdata goaStatistics.csv;
	vdata=ans.data;
	cdfplot(vdata(:,3));hold all;
	cdfplot(vdata(:,5));
	importdata goaGP2GP.csv;
	gdata=ans.data;
	cdfplot(gdata(:,3));`

	fmt.Println(str)

	return

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
type VLmap map[int]VillageInfo

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
			v.distance = radioranges[0:10]
			v.Neighbours = gpids.At(sindx[0:10]...)

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
