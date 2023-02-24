package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/jaswdr/faker"
)

// --- Globals
var fake = faker.New()

// ---  Types
type Place struct {
	Name           string  `json:"name"`
	ZoneIdentifier string  `json:"zoneIdentifier"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

type Zone struct {
	LeftFrontier   float64 `json:"leftFrontier"`
	BottomFrontier float64 `json:"bottomFrontier"`
	RightFrontier  float64 `json:"rightFrontier"`
	TopFrontier    float64 `json:"topFrontier"`
	Identifier     string  `json:"identifier"`
	Number         int     `json:"number"`
}

type ConcurrentPlaces struct {
	sync.RWMutex
	Places []Place
}

type ConcurrentZones struct {
	sync.RWMutex
	Zones []Zone
}

func (c *ConcurrentPlaces) Append(item Place) {
	c.Lock()
	defer c.Unlock()
	c.Places = append(c.Places, item)
}

func (c *ConcurrentZones) Append(item Zone) {
	c.Lock()
	defer c.Unlock()
	c.Zones = append(c.Zones, item)
}

// -- Helper functions

// GetXMLResponse makes a GET request to the given URL and returns the response body as a string
func GetXMLResponse(url string, retried bool) string {
	// Wait 3 seconds before making the request to avoid being blocked
	time.Sleep(3 * time.Second)
	response, err := http.Get(url)

	if err != nil {
		log := fmt.Sprintf("✖ Error making request: %s \n", err)
		color.Red(log)
		return ""
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log := fmt.Sprintf("✖ Status code error: %d \n", response.StatusCode)
		color.Red(log)

		if !retried {
			// Retry after 1 minute to avoid being blocked
			color.Magenta("Retrying in 1 minutes... ")
			time.Sleep(1 * time.Minute)
			return GetXMLResponse(url, true)
		}

		return ""
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		color.Red("✖ Error parsing XML response: ", err, "\n")
		return ""
	}

	return string(data)
}

// GetRandomPlace returns a random place from the given slice
func GetRandomPlace(places *[]Place) (Place, bool) {
	if len(*places) == 0 {
		return Place{}, false
	}

	randomIndex := rand.Intn(len(*places))
	return (*places)[randomIndex], true
}

// GetUniquePlaces returns a slice with the non-duplicated places comparing the names
func GetUniquePlaces(places *[]Place, zones *[]Zone, step float64) []Place {
	uniquePlaces := []Place{}

	for _, place := range *places {
		isUnique := true

		for _, uniquePlace := range uniquePlaces {
			if place.Name == uniquePlace.Name && place.Latitude == uniquePlace.Latitude && place.Longitude == uniquePlace.Longitude {
				isUnique = false
				break
			}
		}

		if isUnique {
			// If the place is unique, add it to the unique places slice
			uniquePlaces = append(uniquePlaces, place)
		} else {
			// Get the zone of the duplicated place and generate a new gym for that zone
			log := fmt.Sprintf("ℹ Duplicated place: %s \n", place.Name)
			color.Blue(log)
			zone, finded := GetZoneByIdentifier(place.ZoneIdentifier, zones)

			if !finded {
				log := fmt.Sprintf("✖ Zone not found: %s \n", place.ZoneIdentifier)
				color.Red(log)
				continue
			}

			// Generate a new place
			randomLat, randomLong := GetRandomPointInZone(zone.LeftFrontier, zone.RightFrontier, zone.BottomFrontier, zone.TopFrontier, step)
			randomName := GetRandomPlaceName()
			newPlace := Place{
				Name:           randomName,
				ZoneIdentifier: place.ZoneIdentifier,
				Latitude:       randomLat,
				Longitude:      randomLong,
			}

			uniquePlaces = append(uniquePlaces, newPlace)

			log = fmt.Sprintf("🎲 Generated random place: %s (%f, %f) \n", randomName, randomLat, randomLong)
			color.Yellow(log)
		}
	}

	return uniquePlaces
}

// GetZoneByIdentifier returns the zone with the given identifier and a boolean indicating if it was found
func GetZoneByIdentifier(identifier string, zones *[]Zone) (Zone, bool) {
	for _, zone := range *zones {
		if zone.Identifier == identifier {
			return zone, true
		}
	}

	return Zone{}, false
}

// GetSortedZones sorts the given zones by top frontier and then by left frontier
func GetSortedZones(zones *[]Zone) []Zone {
	sort.Slice(*zones, func(i, j int) bool {
		// Try to sort by top frontier
		if (*zones)[i].BottomFrontier != (*zones)[j].BottomFrontier {
			return (*zones)[i].BottomFrontier < (*zones)[j].BottomFrontier
		}

		// If top frontiers are equal, sort by left frontier
		return (*zones)[i].LeftFrontier < (*zones)[j].LeftFrontier
	})

	// Add the zone number to each zone
	for i, zone := range *zones {
		zone.Number = i + 1
		(*zones)[i] = zone
	}

	return *zones
}

// SaveStructToFile marshals the given data and saves it to the given file
func SaveStructToFile(data interface{}, fileName string) {
	// Marshal data
	file, err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		log := fmt.Sprintf("✖ Error marshalling data: %s \n", err)
		color.Red(log)
	}

	// Write data to file
	// Crete file on previous directory
	path := "../../data/" + fileName
	err = ioutil.WriteFile(path, file, 0644)

	if err != nil {
		log := fmt.Sprintf("✖ Error writing data to file: %s \n", err)
		color.Red(log)
	}
}

// GetRandomPointInZone returns a random point between the given frontiers
func GetRandomPointInZone(leftFrontier, rightFrontier, bottomFrontier, topFrontier, step float64) (float64, float64) {
	// Reduce the zone to avoid getting points too close to the frontier
	leftFrontier = leftFrontier + step/8
	rightFrontier = rightFrontier - step/8
	bottomFrontier = bottomFrontier + step/8
	topFrontier = topFrontier - step/8

	// Get random point in the zone
	randomLatitude := rand.Float64()*(topFrontier-bottomFrontier) + bottomFrontier
	randomLongitude := rand.Float64()*(rightFrontier-leftFrontier) + leftFrontier

	return randomLatitude, randomLongitude
}

// GetRandomPlaceName returns a random place name
func GetRandomPlaceName() string {
	return fake.Address().StreetName()
}
