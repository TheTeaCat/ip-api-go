# ip-api-go

This is a go module for using [ip-api](https://ip-api.com/) to make [batch queries](https://ip-api.com/docs/api:batch).

With regards to the /batch endpoint, the [ip-api batch endpoint docs](https://ip-api.com/docs/api:batch) state:

> This endpoint is limited to **15** requests per minute from an IP address.

The geolocator in this go module consequently makes a request every 5 seconds (12 times a minute) to safely respect this limit. It is consequently capable of geolocating up to 1200 IPs per minute.



## Usage

`NewGeolocator` should be used to create a `Geolocator` instance. `NewGeolocator` takes one integer argument specifying the size of the queue in the `Geolocator`. If the `Geolocator` runs out of space in its queue it will return an error, `GeolocatorQueueFull`, when you try to call `Locate` with a new IP.

Once you have a `Geolocator` instance, you can call `Locate` on it whenever you want. If the `Geolocator` hasn't queried ip-api for its geolocation yet, and it hasn't been queued, it will be queued. Until the `Geolocator` has queried ip-api for its geolocation, a `LocationNotYetFound` error will be returned. Once the geolocator has queried ip-api and cached the value for the IP you're requesting, it will return a `Geolocation`.

The cache of a `Geolocator` is not automatically cleared. You must call `ClearCache` on it periodically, or when its cache gets too big.



## Example usage

```golang
package main

import (
	"fmt"
	"time"

	geolocator "github.com/TheTeaCat/ip-api-go"
)

func main() {
	const targetIP = "8.8.8.8"
	g := geolocator.NewGeolocator(1000)

	for {
		location, err := g.Locate(targetIP)

		// If any other error occurs, something has gone wrong.
		if err != nil && err.Error() != geolocator.LocationNotYetFound {
			fmt.Printf(
				"Something went wrong, err: %[1]v\n", err.Error(),
			)
			return
		}

		// If LocationNotYetFound occurs, the geolocator hasn't gotten round to processing our request.
		if err != nil && err.Error() == geolocator.LocationNotYetFound {
			fmt.Printf(
				"Still locating %[1]v...\n", targetIP,
			)
			time.Sleep(1 * time.Second) // We wait a second before trying again.
			continue
		}

		fmt.Printf(
			"%[1]v is in (The) %[2]v!\n",
			targetIP, location.Country, err,
		)
		return
	}
}
```



Expected output:

```
Still locating 8.8.8.8...
Still locating 8.8.8.8...
Still locating 8.8.8.8...
Still locating 8.8.8.8...
Still locating 8.8.8.8...
Still locating 8.8.8.8...
8.8.8.8 is in (The) United States!
```

