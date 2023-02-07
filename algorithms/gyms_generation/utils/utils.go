package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"

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
