package collector

import (
	"igdb-database/db"
	"igdb-database/model"
	"log"
	"math"
	"sync"
	"sync/atomic"

	"github.com/bestnite/go-igdb/endpoint"
)

func FetchAndStore[T any](
	e endpoint.EntityEndpoint[T],
) {
	total, err := e.GetLastOneId()
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
				log.Printf("failed to get items: %v", err)
				return
			}

			type IdGetter interface {
				GetId() uint64
			}

			ids := make([]uint64, 0, len(items))
			for _, item := range items {
				if v, ok := any(item).(IdGetter); ok {
					ids = append(ids, v.GetId())
				} else {
					log.Printf("failed to get id from item: %v", err)
					return
				}
			}

			data, err := db.GetItemsByIGDBIDs[T](e.GetEndpointName(), ids)
			if err != nil {
				log.Printf("failed to get items: %v", err)
				return
			}

			newItems := make([]*model.Item[T], 0, len(items))
			for _, item := range items {
				v, ok := any(item).(IdGetter)
				if !ok {
					log.Printf("failed to get id from item: %v", err)
					return
				} else {
					if data[v.GetId()] == nil {
						newItems = append(newItems, model.NewItem(item))
					} else {
						data[v.GetId()].Item = item
						newItems = append(newItems, data[v.GetId()])
					}
				}
			}
			err = db.SaveItems(e.GetEndpointName(), newItems)
			if err != nil {
				log.Printf("failed to save games: %v", err)
				return
			}

			cur := atomic.AddInt32(&finished, 1)
			log.Printf("%s finished: %d/%d", e.GetEndpointName(), cur, totalSteps)
		}(i)
	}
	wg.Wait()
}
