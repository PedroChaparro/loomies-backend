package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/PedroChaparro/loomies-backend-gymsgeneration/utils"
	"github.com/fatih/color"
	"github.com/remeh/sizedwaitgroup"
	"github.com/subchen/go-xmldom"
)

func parseXMLToPlace(xmlPlace *xmldom.Node) utils.Place {
	place := utils.Place{}

	// Get and parse latitude
	placeLat, err := strconv.ParseFloat(xmlPlace.GetAttribute("lat").Value, 64)

	if err != nil {
		color.Red("✖ Error parsing latitude: ", err, "\n")
	}
	place.Latitude = placeLat

	// Get and parse longtitude
	placeLong, err := strconv.ParseFloat(xmlPlace.GetAttribute("lon").Value, 64)

	if err != nil {
		color.Red("✖ Error parsing longitude: ", err, "\n")
	}
	place.Longitude = placeLong

	// Get and parse place name
	for _, child := range xmlPlace.Children {
		if child.GetAttribute("k").Value == "name" {
			place.Name = child.GetAttribute("v").Value
			break
		}
	}

	return place
}

func getBoundedBoxPlaces(left, bottom, right, top float64) []utils.Place {
	// Get API response
	log := fmt.Sprintf("ℹ Getting places between l: %f, b:%f r:%f, t:%f \n", left, bottom, right, top)
	color.Blue(log)

	baseUrl := fmt.Sprintf("http://api.openstreetmap.org/api/0.6/map?bbox=%f,%f,%f,%f", left, bottom, right, top)
	xmlResponse := utils.GetXMLResponse(baseUrl, false)
	if xmlResponse == "" {
		return []utils.Place{}
	}

	// Parse XML response as DOM and get all the <node> elements
	dom := xmldom.Must(xmldom.ParseXML(xmlResponse))
	root := dom.Root
	nodes := root.GetChildren("node")

	// Filter nodes with <tag> and k="name"
	xmlPlaces := []*xmldom.Node{}

	for _, node := range nodes {
		hasTag := false

		for _, child := range node.Children {
			if child.GetAttribute("k").Value == "name" {
				hasTag = true
				break
			}
		}

		if hasTag {
			xmlPlaces = append(xmlPlaces, node)
		}
	}

	// Parse XML nodes to Place objects
	places := []utils.Place{}

	for _, xmlPlace := range xmlPlaces {
		place := parseXMLToPlace(xmlPlace)
		places = append(places, place)
	}

	return places
}

func generatePlaces(minLat, minLong, maxLat, maxLong, step float64) []utils.Place {
	concurrentPlaces := utils.ConcurrentSlice{}
	swg := sizedwaitgroup.New(4)

	for long := minLong; long <= maxLong; long += step {
		for lat := minLat; lat <= maxLat; lat += step {
			swg.Add()

			go func(lat, long, step float64) {
				defer swg.Done()
				boundedPlaces := getBoundedBoxPlaces(lat, long, lat+step, long+step)
				place, success := utils.GetRandomPlace(&boundedPlaces)

				if !success {
					log := fmt.Sprintf("⚠ No place found in bounded box: (%f, %f) and (%f, %f) \n", lat, long, lat+step, long+step)
					color.Yellow(log)
				} else {
					log := fmt.Sprintf("Found place: %s (%f, %f) \n", place.Name, place.Latitude, place.Longitude)
					color.Green(log)
					concurrentPlaces.Append(place)
				}
			}(lat, long, step)

		}
	}

	swg.Wait()
	return concurrentPlaces.Items
}

func main() {
	start := time.Now()
	places := generatePlaces(-73.1000, 6.9629, -73.0320, 7.0500, 0.0035)
	end := time.Now()
	elapsed := end.Sub(start)
	log := fmt.Sprintf("Obtained %d places in %f minutes\n", len(places), elapsed.Minutes())
	color.Green(log)

	// Remove duplicates and save to json file
	uniquePlaces := utils.GetUniquePlaces(&places)
	file, err := json.MarshalIndent(uniquePlaces, "", "  ")
	if err != nil {
		color.Red("✖ Error marshalling places: ", err, "\n")
	}

	err = ioutil.WriteFile("places.json", file, 0644)
	if err != nil {
		color.Red("✖ Error writing places to file: ", err, "\n")
	}
}
