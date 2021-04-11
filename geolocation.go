package geolocator

import "time"

/*Geolocation contains the geolocation data from an ip-api query response */
type Geolocation struct {
	Status        string
	Message       string
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
	geolocation Geolocation
	loaded      bool
	loadedAt    time.Time
	err         error
}

var dummyGeolocations []Geolocation = []Geolocation{
	{
		Status:        "success",
		Message:       "",
		Continent:     "North America",
		ContinentCode: "NA",
		Country:       "United States",
		CountryCode:   "US",
		Region:        "VA",
		RegionName:    "Virginia",
		City:          "Ashburn",
		District:      "",
		Zip:           "20149",
		Lat:           39.03,
		Lon:           -77.5,
		Timezone:      "America/New_York",
		Offset:        -18000,
		Currency:      "USD",
		Isp:           "Google LLC",
		Org:           "Google Public DNS",
		As:            "AS15169 Google LLC",
		Asname:        "GOOGLE",
		Mobile:        false,
		Proxy:         false,
		Hosting:       false,
	},
	{
		Status:        "success",
		Message:       "",
		Continent:     "Europe",
		ContinentCode: "EU",
		Country:       "Ireland",
		CountryCode:   "IE",
		Region:        "L",
		RegionName:    "Leinster",
		City:          "Dublin",
		District:      "",
		Zip:           "D02",
		Lat:           53.3498,
		Lon:           -6.26031,
		Timezone:      "Europe/Dublin",
		Offset:        0,
		Currency:      "EUR",
		Isp:           "Amazon.com, Inc.",
		Org:           "AWS EC2 (eu-west-1)",
		As:            "AS16509 Amazon.com, Inc.",
		Asname:        "AMAZON-02",
		Mobile:        false,
		Proxy:         false,
		Hosting:       false,
	},
	{
		Status:        "success",
		Message:       "",
		Continent:     "Asia",
		ContinentCode: "AS",
		Country:       "Japan",
		CountryCode:   "JP",
		Region:        "27",
		RegionName:    "Ōsaka",
		City:          "Osaka",
		District:      "",
		Zip:           "543-0062",
		Lat:           34.6851,
		Lon:           135.5136,
		Timezone:      "Asia/Tokyo",
		Offset:        32400,
		Currency:      "JPY",
		Isp:           "NTT PC Communications, Inc.",
		Org:           "InfoSphere",
		As:            "AS2514 NTT PC Communications, Inc.",
		Asname:        "INFOSPHERE",
		Mobile:        false,
		Proxy:         false,
		Hosting:       false,
	},
}
