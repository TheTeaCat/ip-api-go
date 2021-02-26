package geolocator

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

const queryURL = "http://ip-api.com/batch?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,offset,currency,isp,org,as,asname,mobile,proxy,hosting,query"

func (g *Geolocator) locateBatch(IPs []string) {
	//responseData will hold all of the responses we get from ip-api
	responseData := make([]Geolocation, 0)

	//err will hold any error that occurs in the process of querying ip-api for this batch of geolocations
	var err error

	//defer processing the batch with the error provided.
	defer g.processBatch(IPs, &responseData, &err)

	//If we're in dev mode then we needn't go further
	if g.dev {
		return
	}

	//Create ip-api query data for POST request.
	queryData, err := json.Marshal(IPs)

	//Make bulk query to ip-api
	requestBody := strings.NewReader(string(queryData))
	r, err := http.Post(queryURL, "application/json", requestBody)
	if err != nil {
		return
	}
	defer r.Body.Close()

	//Unpack results from the ip-api query
	responseRaw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	//Unmarshal the result from the ip-api query
	err = json.Unmarshal(responseRaw, &responseData)
}

func (g *Geolocator) processBatch(IPs []string, geolocations *[]Geolocation, err *error) {
	//We're going to be doing a lot of reads and writes to the cache so we lock it now and just defer the call to Unlock.
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()

	if g.dev {
		// If we're in dev mode we just use dummy locations
		for _, IP := range IPs {
			g.cache[IP].geolocation = dummyGeolocations[rand.Intn(len(dummyGeolocations))]
			g.cache[IP].geolocation.Query = IP
		}
	} else {
		//Otherwise we go over every geolocation and add it to the cache if the query matches a value
		for _, geolocation := range *geolocations {
			_, cachedValExists := g.cache[geolocation.Query]
			if cachedValExists {
				g.cache[geolocation.Query].geolocation = geolocation
			}
		}
	}

	//Update the loaded state and error for every IP in the batch.
	for _, IP := range IPs {
		g.cache[IP].loaded = true
		g.cache[IP].err = *err
	}
}
