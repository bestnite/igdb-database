package main

import (
	"igdb-database/collector"
	"igdb-database/config"
	"igdb-database/db"
	"log"

	"github.com/bestnite/go-igdb"
	"github.com/bestnite/go-igdb/endpoint"
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

	collector.StartWebhookServer(client)
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
