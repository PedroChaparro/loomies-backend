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
)

// ---  Types
type Place struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Zone struct {
	LeftFrontier   float64 `json:"leftFrontier"`
	BottomFrontier float64 `json:"bottomFrontier"`
	RightFrontier  float64 `json:"rightFrontier"`
	TopFrontier    float64 `json:"topFrontier"`
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
func GetUniquePlaces(places *[]Place) []Place {
	uniquePlaces := []Place{}

	for _, place := range *places {
		isUnique := true

		for _, uniquePlace := range uniquePlaces {
			if place.Name == uniquePlace.Name {
				isUnique = false
				break
			}
		}

		if isUnique {
			uniquePlaces = append(uniquePlaces, place)
		}
	}

	return uniquePlaces
}

func GetSortedZones(zones *[]Zone) []Zone {
	sort.Slice(*zones, func(i, j int) bool {
		// Try to sort by top frontier
		if (*zones)[i].BottomFrontier != (*zones)[j].BottomFrontier {
			return (*zones)[i].BottomFrontier < (*zones)[j].BottomFrontier
		}

		// If top frontiers are equal, sort by left frontier
		return (*zones)[i].LeftFrontier < (*zones)[j].LeftFrontier
	})

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
	err = ioutil.WriteFile(fileName, file, 0644)

	if err != nil {
		log := fmt.Sprintf("✖ Error writing data to file: %s \n", err)
		color.Red(log)
	}
}
