package main

import (
	"flag"
	"igdb-database/collector"
	"igdb-database/config"
	"igdb-database/db"
	"log"
	"sync"
	"sync/atomic"

	"github.com/bestnite/go-igdb"
	"github.com/bestnite/go-igdb/endpoint"
	pb "github.com/bestnite/go-igdb/proto"
)

var (
	enableAggregate   = flag.Bool("aggregate", false, "aggregate games")
	enableFetch       = flag.Bool("fetch", false, "fetch data")
	enableReFetch     = flag.Bool("re-fetch", false, "re fetch data even if collection is not empty")
	enableReAggregate = flag.Bool("re-aggregate", false, "re aggregate games even if game_details is not empty")
	enableWebhook     = flag.Bool("webhook", true, "start webhook server")
	onlyRefetchGames  = flag.Bool("only-refetch-games", false, "only refetch games")
)

func main() {
	flag.Parse()

	client := igdb.New(config.C().Twitch.ClientID, config.C().Twitch.ClientSecret)

	if *enableFetch || *enableReFetch {
		log.Printf("fetching data")
		allFetchAndStore(client)
		log.Printf("data fetched")
	}

	Count()

	if *enableAggregate || *enableReAggregate {
		log.Printf("aggregating games")
		aggregateGames()
		log.Printf("games aggregated")
	}

	if *enableWebhook {
		log.Printf("starting webhook server")
		collector.StartWebhookServer(client)
	}
}

func aggregateGames() {
	total, err := db.EstimatedDocumentCount(endpoint.EPGames)
	if err != nil {
		log.Fatalf("failed to count games: %v", err)
	}
	log.Printf("games length: %d", total)

	finished := int64(0)
	wg := sync.WaitGroup{}

	concurrenceNum := 10
	taskOneLoop := int64(500)

	concurrence := make(chan struct{}, concurrenceNum)
	defer close(concurrence)
	for i := int64(0); i < total; i += taskOneLoop {
		concurrence <- struct{}{}
		wg.Add(1)
		go func(i int64) {
			defer func() { <-concurrence }()
			defer wg.Done()
			items, err := db.GetItemsPaginated[pb.Game](endpoint.EPGames, i, taskOneLoop)
			if err != nil {
				log.Fatalf("failed to get games: %v", err)
			}
			isAggregated := make(map[uint64]bool, len(items))
			if !*enableReAggregate {
				isAggregated, err = db.IsGamesAggregated(items)
				if err != nil {
					log.Fatalf("failed to check if games are aggregated: %v", err)
				}
			} else {
				for _, game := range items {
					isAggregated[game.Id] = false
				}
			}
			for _, item := range items {
				if isAggregated[item.Id] {
					p := atomic.AddInt64(&finished, 1)
					log.Printf("game aggregated %d/%d", p, total)
					continue
				}

				game, err := db.ConvertGame(item)
				if err != nil {
					log.Fatalf("failed to convert game: %v", err)
				}

				err = db.SaveGame(game)
				if err != nil {
					log.Fatalf("failed to save game: %v", err)
				}
				p := atomic.AddInt64(&finished, 1)
				log.Printf("game aggregated %d/%d", p, total)
			}
		}(i)
	}
	wg.Wait()
}

func fetchAndStore[T any](
	e endpoint.EntityEndpoint[T],
) {
	if count, err := db.EstimatedDocumentCount(e.GetEndpointName()); (err == nil && count == 0) || *enableReFetch {
		collector.FetchAndStore(e)
	} else if err != nil {
		log.Printf("failed to count items: %v", err)
	}
}

func Count() {
	ids, err := db.GetAllItemsIDs[pb.Theme](endpoint.EPThemes)
	if err != nil {
		log.Fatalf("failed to get all items ids %s: %v", endpoint.EPThemes, err)
	}

	concurrence := make(chan struct{}, 10)
	defer close(concurrence)
	wg := sync.WaitGroup{}

	for _, id := range ids {
		concurrence <- struct{}{}
		wg.Add(1)
		go func(id uint64) {
			defer func() {
				<-concurrence
				wg.Done()
			}()
			count, err := db.CountTheme(id)
			if err != nil {
				log.Fatalf("failed to count theme: %v", err)
			}
			log.Printf("theme %d count: %d", id, count)
			err = db.SaveCount(&db.Count{Theme: id, Count: count})
			if err != nil {
				log.Fatalf("failed to save count: %v", err)
			}
		}(id)
	}

	ids, err = db.GetAllItemsIDs[pb.Genre](endpoint.EPGenres)
	if err != nil {
		log.Fatalf("failed to get all items ids %s: %v", endpoint.EPGenres, err)
	}
	for _, id := range ids {
		concurrence <- struct{}{}
		wg.Add(1)
		go func(id uint64) {
			defer func() {
				<-concurrence
				wg.Done()
			}()
			count, err := db.CountGenre(id)
			if err != nil {
				log.Fatalf("failed to count genre: %v", err)
			}
			log.Printf("genre %d count: %d", id, count)
			err = db.SaveCount(&db.Count{Genre: id, Count: count})
			if err != nil {
				log.Fatalf("failed to save count: %v", err)
			}
		}(id)
	}

	wg.Wait()
}

func allFetchAndStore(client *igdb.Client) {
	if *onlyRefetchGames {
		fetchAndStore(client.Games)
		return
	}

	fetchAndStore(client.AgeRatingCategories)
	fetchAndStore(client.AgeRatingContentDescriptions)
	fetchAndStore(client.AgeRatingContentDescriptionsV2)
	fetchAndStore(client.AgeRatingOrganizations)
	fetchAndStore(client.AgeRatings)
	fetchAndStore(client.AlternativeNames)
	fetchAndStore(client.Artworks)
	fetchAndStore(client.CharacterGenders)
	fetchAndStore(client.CharacterMugShots)
	fetchAndStore(client.Characters)
	fetchAndStore(client.CharacterSpecies)
	fetchAndStore(client.CollectionMemberships)
	fetchAndStore(client.CollectionMembershipTypes)
	fetchAndStore(client.CollectionRelations)
	fetchAndStore(client.CollectionRelationTypes)
	fetchAndStore(client.Collections)
	fetchAndStore(client.CollectionTypes)
	fetchAndStore(client.Companies)
	fetchAndStore(client.CompanyLogos)
	fetchAndStore(client.CompanyStatuses)
	fetchAndStore(client.CompanyWebsites)
	fetchAndStore(client.Covers)
	fetchAndStore(client.DateFormats)
	fetchAndStore(client.EventLogos)
	fetchAndStore(client.EventNetworks)
	fetchAndStore(client.Events)
	fetchAndStore(client.ExternalGames)
	fetchAndStore(client.ExternalGameSources)
	fetchAndStore(client.Franchises)
	fetchAndStore(client.GameEngineLogos)
	fetchAndStore(client.GameEngines)
	fetchAndStore(client.GameLocalizations)
	fetchAndStore(client.GameModes)
	fetchAndStore(client.GameReleaseFormats)
	fetchAndStore(client.GameStatuses)
	fetchAndStore(client.GameTimeToBeats)
	fetchAndStore(client.GameTypes)
	fetchAndStore(client.GameVersionFeatures)
	fetchAndStore(client.GameVersionFeatureValues)
	fetchAndStore(client.GameVersions)
	fetchAndStore(client.GameVideos)
	fetchAndStore(client.Genres)
	fetchAndStore(client.InvolvedCompanies)
	fetchAndStore(client.Keywords)
	fetchAndStore(client.Languages)
	fetchAndStore(client.LanguageSupports)
	fetchAndStore(client.LanguageSupportTypes)
	fetchAndStore(client.MultiplayerModes)
	fetchAndStore(client.NetworkTypes)
	fetchAndStore(client.PlatformFamilies)
	fetchAndStore(client.PlatformLogos)
	fetchAndStore(client.Platforms)
	fetchAndStore(client.PlatformTypes)
	fetchAndStore(client.PlatformVersionCompanies)
	fetchAndStore(client.PlatformVersionReleaseDates)
	fetchAndStore(client.PlatformVersions)
	fetchAndStore(client.PlatformWebsites)
	fetchAndStore(client.PlayerPerspectives)
	fetchAndStore(client.PopularityPrimitives)
	fetchAndStore(client.PopularityTypes)
	fetchAndStore(client.Regions)
	fetchAndStore(client.ReleaseDateRegions)
	fetchAndStore(client.ReleaseDates)
	fetchAndStore(client.ReleaseDateStatuses)
	fetchAndStore(client.Screenshots)
	fetchAndStore(client.Themes)
	fetchAndStore(client.Websites)
	fetchAndStore(client.WebsiteTypes)

	fetchAndStore(client.Games)
}
