// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains a simple command line tool for Geocoding API
// Documentation: https://developers.google.com/maps/documentation/geocoding/
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/kr/pretty"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

var (
	apiKey       = flag.String("key", "", "API Key for using Google Maps API.")
	clientID     = flag.String("client_id", "", "ClientID for Maps for Work API access.")
	signature    = flag.String("signature", "", "Signature for Maps for Work API access.")
	address      = flag.String("address", "", "The street address that you want to geocode, in the format used by the national postal service of the country concerned.")
	components   = flag.String("components", "", "A component filter for which you wish to obtain a geocode.")
	bounds       = flag.String("bounds", "", "The bounding box of the viewport within which to bias geocode results more prominently.")
	language     = flag.String("language", "", "The language in which to return results.")
	region       = flag.String("region", "", "The region code, specified as a ccTLD two-character value.")
	latlng       = flag.String("latlng", "", "The textual latitude/longitude value for which you wish to obtain the closest, human-readable address.")
	resultType   = flag.String("result_type", "", "One or more address types, separated by a pipe (|).")
	locationType = flag.String("location_type", "", "One or more location types, separated by a pipe (|).")
)

func usageAndExit(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	fmt.Println("Flags:")
	flag.PrintDefaults()
	os.Exit(2)
}

func check(err error) {
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
}

var client *maps.Client

type Record struct {
	Indx             int
	FormattedAddress string
	Location         string
	Types            string
	Area             float64
}

func main() {
	flag.Parse()

	bytes, er := ioutil.ReadFile("mykey")
	if er == nil {
		*apiKey = string(bytes)

	}
	var err error
	if *apiKey != "" {
		client, err = maps.NewClient(maps.WithAPIKey(*apiKey))
	} else if *clientID != "" || *signature != "" {
		client, err = maps.NewClient(maps.WithClientIDAndSignature(*clientID, *signature))
	} else {
		usageAndExit("Please specify an API Key, or Client ID and Signature.")
	}
	check(err)

	src, err := os.Open("kkinput.csv")

	check(err)

	fd, err := os.OpenFile("KK.csv", os.O_RDWR|os.O_APPEND, 0660)

	check(err)

	// csvw := csv.NewWriter(fd)
	// writer := csvutil.NewWriter(fd, nil)

	wr := csv.NewWriter(fd)
	rd := csv.NewReader(src)
	rd.Comma = ','
	rd.TrailingComma = true
	indx := 1
	for {

		if indx == 1 {
			indx++
			continue
		}
		record, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(indx, record, len(record))

		// r.Address = "Ambagaratur,Karaikal"
		// r.Address = "Surla, Bicholim"
		// parseComponents(*components, r)
		// parseBounds(*bounds, r)
		// parseLatLng(*latlng, r)
		// parseResultType(*resultType, r)
		// parseLocationType(*locationType, r)

		rowstr := getRecord(record[1], indx)
		log.Printf("%v", rowstr)
		wr.Write(rowstr)
		wr.Flush()
		indx++
	}

	fd.Close()
	// pretty.Println(resp)
}

func getRecord(name string, id int) []string {

	r := &maps.GeocodingRequest{
		Address:  *address,
		Language: *language,
		Region:   *region,
	}

	r.Address = name
	// parseComponents(*components, r)
	// parseBounds(*bounds, r)
	// parseLatLng(*latlng, r)
	// parseResultType(*resultType, r)
	// parseLocationType(*locationType, r)
	log.Println(r)
	resp, err := client.Geocode(context.Background(), r)
	// check(err)
	if err != nil {
		log.Println("Error ", err, name)
	}

	if len(resp) > 1 {
		log.Printf("%s : MORE THAN ONE RESPONSE %d", name, len(resp))
	}
	// var rec Record
	var results []string
	FIELDS := 5
	for i, r := range resp {
		_ = i

		pretty.Println("AddressComponents", r.AddressComponents)
		// fmt.Printf("\n %d : FormattedAddress %#v", i, r.FormattedAddress)
		// pretty.Println("Geometry", r.Geometry)
		// pretty.Printf("\n %d : PlaceID %#v", i, r.PlaceID)
		// pretty.Printf("\n %d : Types %#v", i, r.Types)
		// p1, p2 := r.Geometry.Viewport.NorthEast, r.Geometry.Viewport.SouthWest
		// fmt.Printf("\n Location @ %v : Diagonal Distance : %f km, Area : %v sqkm ", r.Geometry.Location, Distance(p1, p2), Area(r.Geometry.Viewport))
		// pretty.Println("\n======================================= ")
		// var record []string

		// rec.Location = fmt.Sprintf("%s", r.Geometry.Location.String())
		// rec.Types = fmt.Sprintf("%s", r.Types)
		// rec.FormattedAddress = r.FormattedAddress
		// rec.Area = Area(r.Geometry.Viewport)
		// row := csvutil.FormatRow(rec)
		results = append(results, fmt.Sprintf("%d", id))
		results = append(results, r.FormattedAddress)
		results = append(results, strings.Join(r.Types, ","))
		results = append(results, r.Geometry.Location.String())
		results = append(results, fmt.Sprintf("%f", Area(r.Geometry.Viewport)))

		return results
	}
	// rec.Location = "NIL"
	// row := csvutil.FormatRow(rec)
	empty := make([]string, FIELDS)
	empty[0] = fmt.Sprintf("%d", id)
	empty[1] = "NIL"
	return empty
}

func parseComponents(components string, r *maps.GeocodingRequest) {
	if components != "" {
		c := strings.Split(components, "|")
		for _, cf := range c {
			i := strings.Split(cf, ":")
			switch i[0] {
			case "route":
				r.Components[maps.ComponentRoute] = i[1]
			case "locality":
				r.Components[maps.ComponentLocality] = i[1]
			case "administrative_area":
				r.Components[maps.ComponentAdministrativeArea] = i[1]
			case "postal_code":
				r.Components[maps.ComponentPostalCode] = i[1]
			case "country":
				r.Components[maps.ComponentCountry] = i[1]
			}
		}
	}
}

func parseBounds(bounds string, r *maps.GeocodingRequest) {
	if bounds != "" {
		b := strings.Split(bounds, "|")
		sw := strings.Split(b[0], ",")
		ne := strings.Split(b[1], ",")

		swLat, err := strconv.ParseFloat(sw[0], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}
		swLng, err := strconv.ParseFloat(sw[1], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}
		neLat, err := strconv.ParseFloat(ne[0], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}
		neLng, err := strconv.ParseFloat(ne[1], 64)
		if err != nil {
			log.Fatalf("Couldn't parse bounds: %#v", err)
		}

		r.Bounds = &maps.LatLngBounds{
			NorthEast: maps.LatLng{Lat: neLat, Lng: neLng},
			SouthWest: maps.LatLng{Lat: swLat, Lng: swLng},
		}
	}
}

func parseLatLng(latlng string, r *maps.GeocodingRequest) {
	if latlng != "" {
		l := strings.Split(latlng, ",")
		lat, err := strconv.ParseFloat(l[0], 64)
		if err != nil {
			log.Fatalf("Couldn't parse latlng: %#v", err)
		}
		lng, err := strconv.ParseFloat(l[1], 64)
		if err != nil {
			log.Fatalf("Couldn't parse latlng: %#v", err)
		}
		r.LatLng = &maps.LatLng{
			Lat: lat,
			Lng: lng,
		}
	}
}

func parseResultType(resultType string, r *maps.GeocodingRequest) {
	if resultType != "" {
		r.ResultType = strings.Split(resultType, "|")
	}
}

func parseLocationType(locationType string, r *maps.GeocodingRequest) {
	if locationType != "" {
		for _, l := range strings.Split(locationType, "|") {
			switch l {
			case "ROOFTOP":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyRooftop)
			case "RANGE_INTERPOLATED":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyRangeInterpolated)
			case "GEOMETRIC_CENTER":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyGeometricCenter)
			case "APPROXIMATE":
				r.LocationType = append(r.LocationType, maps.GeocodeAccuracyApproximate)
			}
		}

	}
}

func Distance(p1, p2 maps.LatLng) float64 {

	r := 6371.0 // approx. radius of earth in km
	lat1Radians := (p1.Lat * math.Pi) / 180.0
	lon1Radians := (p1.Lng * math.Pi) / 180.0
	lat2Radians := (p2.Lat * math.Pi) / 180.0
	lon2Radians := (p2.Lng * math.Pi) / 180.0
	d := r * math.Acos(math.Cos(lat1Radians)*math.Cos(lat2Radians)*math.Cos(lon2Radians-lon1Radians)+(math.Sin(lat1Radians)*math.Sin(lat2Radians)))
	return d
}

func Area(bound maps.LatLngBounds) float64 {

	// lat1-lon1 is the upper-left corner, lat2-lon2 is the lower-right
	p1, p2 := bound.NorthEast, bound.SouthWest
	p3 := p1
	p4 := p1
	p3.Lat = p2.Lat
	p4.Lng = p2.Lng
	height := Distance(p1, p3)
	width := Distance(p1, p4)
	// fmt.Println("Width, Height ", height, width)
	return height * width

}
