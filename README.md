# ip-api-go

This is a go module for using ip-api to make [batch queries](https://ip-api.com/docs/api:batch).

With regards to the /batch endpoint, the ip-api docs state:

> This endpoint is limited to **15** requests per minute from an IP address.

The geolocator in this go module consequently makes a request every 5 seconds (12 times a minute) to safely respect this limit.



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

	// Initialise a new geolocator
	g := geolocator.NewGeolocator(1000)

	for {
		// Attempt to get the geolocation of 8.8.8.8
		location, err := g.Locate(targetIP)

		// Log that we're still trying
		fmt.Println("Attempting to geolocate 8.8.8.8...")

		// If the location is not nil, or the error is nil, we're done, so we log the result.
		if location != nil || err == nil {
			fmt.Printf(
				"Found %[1]v! \nLocation: \n%[2]v \nerr: \n%[3]v \n",
				targetIP, location, err,
			)
			return
		}

		// Otherwise, we wait a second before trying again
		time.Sleep(1 * time.Second)
	}
}
```



Expected output:

```
Attempting to geolocate 8.8.8.8...
Attempting to geolocate 8.8.8.8...
Attempting to geolocate 8.8.8.8...
Attempting to geolocate 8.8.8.8...
Attempting to geolocate 8.8.8.8...
Attempting to geolocate 8.8.8.8...
Attempting to geolocate 8.8.8.8...
Found 8.8.8.8! 
Location: 
&{success North America NA United States US VA Virginia Ashburn  20149 39.03 -77.5 America/New_York -18000 USD Google LLC Google Public DNS AS15169 Google LLC GOOGLE false false true 8.8.8.8} 
err: 
<nil>
```

