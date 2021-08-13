package geolocator

import (
	"errors"
	"sync"
	"time"
)

//Geolocator holds a cache and a queue of IPs for which the geolocation has been requested.
type Geolocator struct {
	cache         map[string]*cachedGeolocation
	cacheMutex    *sync.RWMutex
	queueIncoming chan string // gatekeeper for the main ring buffer (ouroboros).
	queueOutgoing chan string // the main queue is a ring buffer.
	dev           bool        //if dev is true, we just use dummy locations.
	/*postBatchCallback is called after each batch was processed with the number
	of IPs that were in that batch (batchSize), and the time since the last batch
	was located (sinceLastBatchLocated). */
	postBatchCallback *func(batchSize int, sinceLastBatchLocated time.Duration)
}

/*NewGeolocator takes an int which specifies the size of the queue the geolocator uses to hold IP location requests. It
is emptied at a rate of up to 100 IPs every 5 seconds, so the queue must be sized sensibly according to your application
in order to mitigate the risk of it overfilling. */
func NewGeolocator(queueCap int, dev bool, postBatchCallback *func(batchSize int, sinceLastBatchLocated time.Duration)) *Geolocator {
	//Create a geolocator
	g := &Geolocator{
		cache:             make(map[string]*cachedGeolocation),
		cacheMutex:        &sync.RWMutex{},
		queueIncoming:     make(chan string, 100),
		queueOutgoing:     make(chan string, queueCap),
		dev:               dev,
		postBatchCallback: postBatchCallback,
	}

	//Start the geolocator (queries ip-api periodically)
	go g.start()

	// Start the ring queue.
	go g.ouroboros()

	//Return the geolocator instance
	return g
}

/*CacheSize simply returns the current size of the cache. Expected to be used to determine when to call ClearCache, or
for logging purposes. */
func (g *Geolocator) CacheSize() int {
	g.cacheMutex.RLock()
	defer g.cacheMutex.RUnlock()
	return len(g.cache)
}

/*QueueSize simply returns the current size of the queue. Expected to be used for logging purposes. */
func (g *Geolocator) QueueSize() int {
	return len(g.queueOutgoing) + len(g.queueIncoming)
}

/*Proxies returns the number of geolocations currently in the geolocator cache flagged as proxies by ip-api*/
func (g *Geolocator) Proxies() int {
	g.cacheMutex.RLock()
	defer g.cacheMutex.RUnlock()
	n := 0
	for _, cachedgeo := range g.cache {
		if cachedgeo.loaded && cachedgeo.geolocation.Proxy {
			n++
		}
	}
	return n
}

/*Hosts returns the number of geolocations currently in the geolocator cache flagged as hosts by ip-api*/
func (g *Geolocator) Hosts() int {
	g.cacheMutex.RLock()
	defer g.cacheMutex.RUnlock()
	n := 0
	for _, cachedgeo := range g.cache {
		if cachedgeo.loaded && cachedgeo.geolocation.Hosting {
			n++
		}
	}
	return n
}

//Delete removes a geolocation from the cache
func (g *Geolocator) Delete(IP string) bool {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	cachedVal, cached := g.cache[IP]
	if cached && cachedVal.loaded {
		delete(g.cache, IP)
	}
	return cached
}

//Prune removes all the cached values older than a provided duration
func (g *Geolocator) Prune(maxAge time.Duration) {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	for IP, cachedVal := range g.cache {
		if time.Since(cachedVal.loadedAt) > maxAge {
			delete(g.cache, IP)
		}
	}
}

//Locate takes an IP and returns a Geolocation. If it's not yet found, it will return nil and an error.
func (g *Geolocator) Locate(IP string) (*Geolocation, error) {
	//Check for a cached value first!
	g.cacheMutex.RLock()
	cachedVal, cached := g.cache[IP]
	g.cacheMutex.RUnlock()

	/* If the value doesn't even have a placeholder in the cache, we enqueue it. */
	if !cached {
		err := g.enqueue(IP) //This may err if the queue is full.
		if err != nil {
			return nil, err
		}
		//The location will not immediately be found so we return a LocationNotYetFound error.
		return nil, errors.New(LocationNotYetFound)
	}

	//If the value hasn't yet been loaded then we still return a LocationNotYetFound error.
	if !cachedVal.loaded {
		return nil, errors.New(LocationNotYetFound)
	}

	//Once it's loaded, we return the geolocation and error straight from the cachedVal (either could still be nil)
	return &cachedVal.geolocation, cachedVal.err
}

func (g *Geolocator) enqueue(IP string) error {
	//First, we add an empty placeholder value to the cache to mark it as requested...
	g.cacheMutex.Lock()
	g.cache[IP] = &cachedGeolocation{
		geolocation: Geolocation{},
		loaded:      false,
		err:         nil,
	}
	g.cacheMutex.Unlock()

	//...then add the IP to the queue.
	select {
	case g.queueIncoming <- IP:
		return nil
	default:
		//If the queue is full, we return the relevant error.
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
			case IP, ok := <-g.queueOutgoing:
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
		sinceLastBatch := time.Since(lastLocateCall)
		if (len(batchToLocate) == 100 || len(g.queueOutgoing) < 10) && sinceLastBatch.Seconds() >= 5 {
			g.locateBatch(batchToLocate)
			//Call our callback if we have one
			if g.postBatchCallback != nil {
				(*g.postBatchCallback)(len(batchToLocate), sinceLastBatch)
			}
			//Remember to update the time of the last ip-api call to time.Now() and clear the batch.
			lastLocateCall = time.Now()
			batchToLocate = make([]string, 0)
		}
	}
}

func (g *Geolocator) ouroboros() {
	for val := range g.queueIncoming {
		select {
		case g.queueOutgoing <- val:
			// could add to ring buffer, do nothing.
		default:
			// outgoing ring full, drop first item.
			<-g.queueOutgoing
			g.queueOutgoing <- val
		}
	}
}
