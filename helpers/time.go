package helpers

import (
	"log"
	"os"
	"time"
)

// LoadLocation loads the specific location
func LoadLocation() (*time.Location, error) {
	return time.LoadLocation(os.Getenv("LOCATION"))
}

// Now returns time.Time objects relying on the specific location
func Now() time.Time {
	l, err := LoadLocation()
	if err != nil {
		log.Fatalln(err)
	}

	return time.Now().In(l)
}
