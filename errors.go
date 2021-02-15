package geolocator

/*LocationNotYetFound is returned as an error when the location requested has
been queued to be geolocated, but it has not yet been geolocated. */
const LocationNotYetFound = "The IP requested has not yet been geolocated"

/*GeolocatorQueueFull is returned as an error if the Geolocator's queue has
become saturated (and consequently no more IPs can be queued to be located)*/
const GeolocatorQueueFull = "The geolocator queue is full"

/*GeolocatorStopped is used as an error to panic if the Geolocator's locatorLoop
stops unexpectedly*/
const GeolocatorStopped = "The geolocator stopped unexpectedly"
