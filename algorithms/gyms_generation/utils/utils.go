package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
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

type ConcurrentSlice struct {
	sync.RWMutex
	Items []Place
}

func (c *ConcurrentSlice) Append(item Place) {
	c.Lock()
	defer c.Unlock()
	c.Items = append(c.Items, item)
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
