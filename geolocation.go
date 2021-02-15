package geolocator

/*Geolocation contains the geolocation data from an ip-api query response */
type Geolocation struct {
	Status        string
	Continent     string
	ContinentCode string
	Country       string
	CountryCode   string
	Region        string
	RegionName    string
	City          string
	District      string
	Zip           string
	Lat           float64
	Lon           float64
	Timezone      string
	Offset        int
	Currency      string
	Isp           string
	Org           string
	As            string
	Asname        string
	Mobile        bool
	Proxy         bool
	Hosting       bool
	Query         string
}

/*cachedGeolocation holds a geolocation when it's fetched from ip-api. Until then, loaded will be false. The geolocation
may remain nil if ip-api fails, in which case loaded will be true and err will contain the relevant error. */
type cachedGeolocation struct {
	geolocation *Geolocation
	loaded      bool
	err         error
}
