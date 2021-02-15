package geolocator

import (
	"errors"
	"sync"
	"time"
)

//Geolocator holds a cache and a queue of IPs for which the geolocation has been requested.
type Geolocator struct {
	cache      map[string]*cachedGeolocation
	cacheMutex *sync.Mutex
	queue      chan string
}

/*NewGeolocator takes an int which specifies the size of the queue the geolocator uses to hold IP location requests. It
is emptied at a rate of up to 100 IPs every 5 seconds, so the queue must be sized sensibly according to your application
in order to mitigate the risk of it overfilling. */
func NewGeolocator(queueCap int) *Geolocator {
	// Create a geolocator
	g := &Geolocator{
		cache:      make(map[string]*cachedGeolocation),
		cacheMutex: &sync.Mutex{},
		queue:      make(chan string, queueCap),
	}

	// Start the geolocator (queries ip-api periodically)
	go g.start()

	// Return the geolocator instance
	return g
}

/*ClearCache clears the cache of the geolocator. It would be prudent to call this at regular intervals, or when g.cache
gets big, to avoid a memory leak. */
func (g *Geolocator) ClearCache() {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	g.cache = make(map[string]*cachedGeolocation)
}

//Locate takes an IP and returns a Geolocation. If it's not yet found, it will return nil and an error.
func (g *Geolocator) Locate(IP string) (*Geolocation, error) {
	// Check for a cached value first!
	g.cacheMutex.Lock()
	cachedVal, cached := g.cache[IP]
	g.cacheMutex.Unlock()

	/* If the value doesn't even have a placeholder in the cache, we enqueue it. */
	if !cached {
		err := g.enqueue(IP) // This may err if the queue is full.
		if err != nil {
			return nil, err
		}
		// The location will not immediately be found so we return a LocationNotYetFound error.
		return nil, errors.New(LocationNotYetFound)
	}

	// If the value hasn't yet been loaded then we still return a LocationNotYetFound error.
	if !cachedVal.loaded {
		return nil, errors.New(LocationNotYetFound)
	}

	// Once it's loaded, we return the geolocation and error straight from the cachedVal (either could still be nil)
	return cachedVal.geolocation, cachedVal.err
}

func (g *Geolocator) enqueue(IP string) error {
	// First, we add an empty placeholder value to the cache to mark it as requested...
	g.cacheMutex.Lock()
	g.cache[IP] = &cachedGeolocation{
		geolocation: nil,
		loaded:      false,
		err:         nil,
	}
	g.cacheMutex.Unlock()

	// ...then add the IP to the queue.
	select {
	case g.queue <- IP:
		return nil
	default:
		// If the queue is full, we return the relevant error.
		return errors.New(GeolocatorQueueFull)
	}
}

func (g *Geolocator) start() {
	//batchToLocate holds a batch of IPs to be located.
	batchToLocate := make([]string, 0)

	//lastLocateCall holds the time of the last call to ip-api's /batch endpoint.
	lastLocateCall := time.Now()

	for {
		//If IPsToLocate isn't full, we check if the queue channel has a value ready to add to the batch
		if len(batchToLocate) < 100 {
			select {
			case IP, ok := <-g.queue:
				if ok {
					batchToLocate = append(batchToLocate, IP)
					continue
				} else {
					panic(errors.New(GeolocatorStopped))
				}
			default:
			}
		}

		/*If it's been at least 5 seconds since the last call to ip-api, and there are IPs ready to locate, then we make the
		locate the batch.*/
		if len(batchToLocate) > 0 && time.Now().Sub(lastLocateCall).Seconds() >= 5 {
			g.locateBatch(batchToLocate)
			//Remember to update the time of the last ip-api call to time.Now() and clear the batch.
			lastLocateCall = time.Now()
			batchToLocate = make([]string, 0)
		}
	}
}
