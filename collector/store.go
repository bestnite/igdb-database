package collector

import (
	"igdb-database/db"
	"log"
	"math"
	"sync"
	"sync/atomic"

	"github.com/bestnite/go-igdb/endpoint"
)

func FetchAndStore[T any](
	e endpoint.EntityEndpoint[T],
) {
	total, err := e.Count()
	if err != nil {
		log.Fatalf("failed to get %s length: %v", e.GetEndpointName(), err)
	}
	log.Printf("%s length: %d", e.GetEndpointName(), total)
	wg := sync.WaitGroup{}
	concurrence := make(chan struct{}, 3)
	defer close(concurrence)

	totalSteps := int(math.Ceil(float64(total) / 500))
	finished := int32(0)

	for i := 0; i < int(total); i += 500 {
		wg.Add(1)
		concurrence <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-concurrence }()

			items, err := e.Paginated(uint64(i), 500)
			if err != nil {
				log.Printf("failed to get items from igdb %s: %v", e.GetEndpointName(), err)
				return
			}

			err = db.SaveItems(e.GetEndpointName(), items)
			if err != nil {
				log.Printf("failed to save games %s: %v", e.GetEndpointName(), err)
				return
			}

			cur := atomic.AddInt32(&finished, 1)
			log.Printf("%s finished: %d/%d", e.GetEndpointName(), cur, totalSteps)
		}(i)
	}
	wg.Wait()
}
