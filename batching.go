package geolocator

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const queryURL = "http://ip-api.com/batch?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,offset,currency,isp,org,as,asname,mobile,proxy,hosting,query"

func (g *Geolocator) locateBatch(IPs []string) {
	//err will hold any error that occurs in the process of querying ip-api for this batch of geolocations
	var err error

	//If we're in dev mode then we needn't go further
	if g.dev {
		g.processBatch(IPs, make([]Geolocation, 0), err)
		return
	}

	//Create ip-api query data for POST request.
	queryData, err := json.Marshal(IPs)

	//Make bulk query to ip-api
	requestBody := strings.NewReader(string(queryData))
	r, err := http.Post(queryURL, "application/json", requestBody)

	//If err, process the batch by applying the error to all the cached vals
	if err != nil {
		g.processBatch(IPs, make([]Geolocation, 0), err)
		return
	}

	//Remember to close our request body!
	defer r.Body.Close()

	//Unpack results from the ip-api query
	responseRaw, err := ioutil.ReadAll(r.Body)
	//If err, process the batch by applying the error to all the cached vals
	if err != nil {
		g.processBatch(IPs, make([]Geolocation, 0), err)
	}

	//responseData will hold all of the responses we get from ip-api
	responseData := make([]Geolocation, 0)

	//Unmarshal the result from the ip-api query
	err = json.Unmarshal(responseRaw, &responseData)

	//process the batch with the error provided.
	g.processBatch(IPs, responseData, err)
}

func (g *Geolocator) processBatch(IPs []string, geolocations []Geolocation, err error) {
	//We're going to be doing a lot of reads and writes to the cache so we lock it now and just defer the call to Unlock.
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()

	//Add all the geolocations to the cache
	if g.dev {
		// If we're in dev mode we just use dummy locations
		for _, IP := range IPs {
			g.cache[IP].geolocation = dummyGeolocations[rand.Intn(len(dummyGeolocations))]
			g.cache[IP].geolocation.Query = IP
		}
	} else {
		//Otherwise we go over every geolocation and add it to the cache if the query matches a value
		for _, geolocation := range geolocations {
			_, cachedValExists := g.cache[geolocation.Query]
			if cachedValExists {
				g.cache[geolocation.Query].geolocation = geolocation
			}
		}
	}

	//The time we'll use for the loadedAt fields of all the geolocations
	loadedAt := time.Now()

	for _, IP := range IPs {
		//This may be nil if it's been cleared by a prune.
		if g.cache[IP] != nil {
			g.cache[IP].err = err
			g.cache[IP].loadedAt = loadedAt
			g.cache[IP].loaded = true
		}
	}
}
