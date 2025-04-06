package main

import (
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

func main() {
	client := igdb.New(config.C().Twitch.ClientID, config.C().Twitch.ClientSecret)

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
	fetchAndStore(client.Games)
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

	log.Printf("aggregating games")
	aggregateGames()
	log.Printf("games aggregated")

	log.Printf("starting webhook server")
	collector.StartWebhookServer(client)
}

func aggregateGames() {
	total, err := db.CountItems(endpoint.EPGames)
	if err != nil {
		log.Fatalf("failed to count games: %v", err)
	}

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
			items, err := db.GetItemsPagnated[pb.Game](endpoint.EPGames, i, taskOneLoop)
			if err != nil {
				log.Fatalf("failed to get games: %v", err)
			}
			games := make([]*pb.Game, 0, len(items))
			for _, item := range items {
				games = append(games, item.Item)
			}
			isAggregated, err := db.IsGamesAggregated(games)
			if err != nil {
				log.Fatalf("failed to check if games are aggregated: %v", err)
			}
			for _, item := range items {
				if isAggregated[item.Item.Id] {
					p := atomic.AddInt64(&finished, 1)
					log.Printf("game aggregated %d/%d", p, total)
					continue
				}
				if err != nil {
					log.Fatalf("failed to check if game is aggregated: %v", err)
				}

				game, err := db.ConvertGame(item.Item)
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
	if count, err := db.CountItems(e.GetEndpointName()); err == nil && count == 0 {
		collector.FetchAndStore(e)
	} else if err != nil {
		log.Printf("failed to count items: %v", err)
	}
}
