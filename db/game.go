package db

import (
	"context"
	"errors"
	"fmt"
	"igdb-database/model"
	"time"

	"github.com/bestnite/go-igdb"
	"github.com/bestnite/go-igdb/endpoint"
	pb "github.com/bestnite/go-igdb/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func IsGamesAggregated(games []*pb.Game) (map[uint64]bool, error) {
	ids := make([]uint64, 0, len(games))
	for _, game := range games {
		ids = append(ids, game.Id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(games))*200*time.Millisecond)
	defer cancel()
	cursor, err := GetInstance().GameCollection.Find(ctx, bson.M{"id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}

	res := make(map[uint64]bool, len(games))
	var g []*model.Game
	err = cursor.All(ctx, &g)
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}
	for _, game := range g {
		res[game.Id] = true
	}

	return res, nil
}

func SaveGame(game *model.Game) error {
	filter := bson.M{"id": game.Id}
	update := bson.M{"$set": game}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().GameCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func SaveGames(games []*model.Game) error {
	updateModel := make([]mongo.WriteModel, 0, len(games))
	for _, game := range games {
		updateModel = append(updateModel, mongo.NewUpdateOneModel().SetFilter(bson.M{"id": game.Id}).SetUpdate(bson.M{"$set": game}).SetUpsert(true))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second+time.Duration(len(games))*200*time.Millisecond)
	defer cancel()
	_, err := GetInstance().GameCollection.BulkWrite(ctx, updateModel)
	if err != nil {
		return err
	}
	return nil
}

type IdGetter interface {
	GetId() uint64
}

func ConvertGame(game *pb.Game, client *igdb.Client) (*model.Game, error) {
	res := &model.Game{}

	if game == nil {
		return nil, fmt.Errorf("game is nil")
	}

	res.Id = game.Id

	ageRatingsIds := make([]uint64, 0, len(game.AgeRatings))
	for _, g := range game.AgeRatings {
		ageRatingsIds = append(ageRatingsIds, g.Id)
	}
	ageRatings, err := GetItemsByIds[pb.AgeRating](endpoint.EPAgeRatings, ageRatingsIds)
	if err != nil {
		return nil, err
	}
	if len(ageRatings) != len(ageRatingsIds) {
		missingIds := make([]uint64, 0, len(ageRatingsIds)-len(ageRatings))
		for _, id := range ageRatingsIds {
			found := false
			for _, item := range ageRatings {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		ageRatings, err = client.AgeRatings.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		ageRatings = append(ageRatings, ageRatings...)
		err = SaveItems(endpoint.EPAgeRatings, ageRatings)
		if err != nil {
			return nil, err
		}
	}
	res.AgeRatings = ageRatings

	res.AggregatedRating = game.AggregatedRating
	res.AggregatedRatingCount = game.AggregatedRatingCount

	alternativeNameIds := make([]uint64, 0, len(game.AlternativeNames))
	for _, g := range game.AlternativeNames {
		alternativeNameIds = append(alternativeNameIds, g.Id)
	}
	alternativeNames, err := GetItemsByIds[pb.AlternativeName](endpoint.EPAlternativeNames, alternativeNameIds)
	if err != nil {
		return nil, err
	}
	if len(alternativeNames) != len(alternativeNameIds) {
		missingIds := make([]uint64, 0, len(alternativeNameIds)-len(alternativeNames))
		for _, id := range alternativeNameIds {
			found := false
			for _, item := range alternativeNames {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		alternativeNames, err = client.AlternativeNames.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		alternativeNames = append(alternativeNames, alternativeNames...)
		err = SaveItems(endpoint.EPAlternativeNames, alternativeNames)
		if err != nil {
			return nil, err
		}
	}
	res.AlternativeNames = alternativeNames

	ArtworkIds := make([]uint64, 0, len(game.Artworks))
	for _, g := range game.Artworks {
		ArtworkIds = append(ArtworkIds, g.Id)
	}
	artworks, err := GetItemsByIds[pb.Artwork](endpoint.EPArtworks, ArtworkIds)
	if err != nil {
		return nil, err
	}
	if len(artworks) != len(ArtworkIds) {
		missingIds := make([]uint64, 0, len(ArtworkIds)-len(artworks))
		for _, id := range ArtworkIds {
			found := false
			for _, item := range artworks {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		artworks, err = client.Artworks.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		artworks = append(artworks, artworks...)
		err = SaveItems(endpoint.EPArtworks, artworks)
		if err != nil {
			return nil, err
		}
	}
	res.Artworks = artworks

	bundlesIds := make([]uint64, 0, len(game.Bundles))
	for _, g := range game.Bundles {
		bundlesIds = append(bundlesIds, g.Id)
	}
	res.Bundles = bundlesIds

	if game.Cover != nil {
		coverId := game.Cover.Id
		cover, err := GetItemById[pb.Cover](endpoint.EPCovers, coverId)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				cover, err = client.Covers.GetByID(coverId)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		if cover != nil {
			res.Cover = cover
		}
	}

	res.CreatedAt = game.CreatedAt

	dlcsIds := make([]uint64, 0, len(game.Dlcs))
	for _, g := range game.Dlcs {
		dlcsIds = append(dlcsIds, g.Id)
	}
	res.Dlcs = dlcsIds

	expansionsIds := make([]uint64, 0, len(game.Expansions))
	for _, g := range game.Expansions {
		expansionsIds = append(expansionsIds, g.Id)
	}
	res.Expansions = expansionsIds

	externalGameIds := make([]uint64, 0, len(game.ExternalGames))
	for _, g := range game.ExternalGames {
		externalGameIds = append(externalGameIds, g.Id)
	}
	externalGames, err := GetItemsByIds[pb.ExternalGame](endpoint.EPExternalGames, externalGameIds)
	if err != nil {
		return nil, err
	}
	if len(externalGames) != len(externalGameIds) {
		missingIds := make([]uint64, 0, len(externalGameIds)-len(externalGames))
		for _, id := range externalGameIds {
			found := false
			for _, item := range externalGames {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		externalGames, err = client.ExternalGames.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		externalGames = append(externalGames, externalGames...)
		err = SaveItems(endpoint.EPExternalGames, externalGames)
		if err != nil {
			return nil, err
		}
	}
	res.ExternalGames = externalGames

	res.FirstReleaseDate = game.FirstReleaseDate

	res.Franchise = nil

	if game.Franchise != nil {
		franchiseId := game.Franchise.Id
		franchise, err := GetItemById[pb.Franchise](endpoint.EPFranchises, franchiseId)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				franchise, err = client.Franchises.GetByID(franchiseId)
				if err != nil {
					return nil, err
				}
				err = SaveItem(endpoint.EPFranchises, franchise)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		if franchise != nil {
			res.Franchise = franchise
		}
	}

	franchiseIds := make([]uint64, 0, len(game.Franchises))
	for _, g := range game.Franchises {
		franchiseIds = append(franchiseIds, g.Id)
	}
	franchises, err := GetItemsByIds[pb.Franchise](endpoint.EPFranchises, franchiseIds)
	if err != nil {
		return nil, err
	}
	if len(franchises) != len(franchiseIds) {
		missingIds := make([]uint64, 0, len(franchiseIds)-len(franchises))
		for _, id := range franchiseIds {
			found := false
			for _, item := range franchises {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		franchises, err = client.Franchises.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		franchises = append(franchises, franchises...)
		err = SaveItems(endpoint.EPFranchises, franchises)
		if err != nil {
			return nil, err
		}
	}
	res.Franchises = franchises

	gameEngineIds := make([]uint64, 0, len(game.GameEngines))
	for _, g := range game.GameEngines {
		gameEngineIds = append(gameEngineIds, g.Id)
	}
	gameEngines, err := GetItemsByIds[pb.GameEngine](endpoint.EPGameEngines, gameEngineIds)
	if err != nil {
		return nil, err
	}
	if len(gameEngines) != len(gameEngineIds) {
		missingIds := make([]uint64, 0, len(gameEngineIds)-len(gameEngines))
		for _, id := range gameEngineIds {
			found := false
			for _, item := range gameEngines {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		gameEngines, err = client.GameEngines.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		gameEngines = append(gameEngines, gameEngines...)
		err = SaveItems(endpoint.EPGameEngines, gameEngines)
		if err != nil {
			return nil, err
		}
	}
	res.GameEngines = gameEngines

	gameModeIds := make([]uint64, 0, len(game.GameModes))
	for _, g := range game.GameModes {
		gameModeIds = append(gameModeIds, g.Id)
	}
	gameModes, err := GetItemsByIds[pb.GameMode](endpoint.EPGameModes, gameModeIds)
	if err != nil {
		return nil, err
	}
	if len(gameModes) != len(gameModeIds) {
		missingIds := make([]uint64, 0, len(gameModeIds)-len(gameModes))
		for _, id := range gameModeIds {
			found := false
			for _, item := range gameModes {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		gameModes, err = client.GameModes.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		gameModes = append(gameModes, gameModes...)
		err = SaveItems(endpoint.EPGameModes, gameModes)
		if err != nil {
			return nil, err
		}
	}
	res.GameModes = gameModes

	genreIds := make([]uint64, 0, len(game.Genres))
	for _, g := range game.Genres {
		genreIds = append(genreIds, g.Id)
	}
	genres, err := GetItemsByIds[pb.Genre](endpoint.EPGenres, genreIds)
	if err != nil {
		return nil, err
	}
	if len(genres) != len(genreIds) {
		missingIds := make([]uint64, 0, len(genreIds)-len(genres))
		for _, id := range genreIds {
			found := false
			for _, item := range genres {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		genres, err = client.Genres.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genres...)
		err = SaveItems(endpoint.EPGenres, genres)
		if err != nil {
			return nil, err
		}
	}
	res.Genres = genres

	res.Hypes = game.Hypes

	involvedCompanyIds := make([]uint64, 0, len(game.InvolvedCompanies))
	for _, g := range game.InvolvedCompanies {
		involvedCompanyIds = append(involvedCompanyIds, g.Id)
	}
	involvedCompanies, err := GetItemsByIds[pb.InvolvedCompany](endpoint.EPInvolvedCompanies, involvedCompanyIds)
	if err != nil {
		return nil, err
	}
	if len(involvedCompanies) != len(involvedCompanyIds) {
		missingIds := make([]uint64, 0, len(involvedCompanyIds)-len(involvedCompanies))
		for _, id := range involvedCompanyIds {
			found := false
			for _, item := range involvedCompanies {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		involvedCompanies, err = client.InvolvedCompanies.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		involvedCompanies = append(involvedCompanies, involvedCompanies...)
		err = SaveItems(endpoint.EPInvolvedCompanies, involvedCompanies)
		if err != nil {
			return nil, err
		}
	}
	res.InvolvedCompanies = involvedCompanies

	keywordIds := make([]uint64, 0, len(game.Keywords))
	for _, g := range game.Keywords {
		keywordIds = append(keywordIds, g.Id)
	}
	keyword, err := GetItemsByIds[pb.Keyword](endpoint.EPKeywords, keywordIds)
	if err != nil {
		return nil, err
	}
	if len(keyword) != len(keywordIds) {
		missingIds := make([]uint64, 0, len(keywordIds)-len(keyword))
		for _, id := range keywordIds {
			found := false
			for _, item := range keyword {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		keyword, err = client.Keywords.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		keyword = append(keyword, keyword...)
		err = SaveItems(endpoint.EPKeywords, keyword)
		if err != nil {
			return nil, err
		}
	}
	res.Keywords = keyword

	multiplayerModeIds := make([]uint64, 0, len(game.MultiplayerModes))
	for _, g := range game.MultiplayerModes {
		multiplayerModeIds = append(multiplayerModeIds, g.Id)
	}
	multiplayerModes, err := GetItemsByIds[pb.MultiplayerMode](endpoint.EPMultiplayerModes, multiplayerModeIds)
	if err != nil {
		return nil, err
	}
	if len(multiplayerModes) != len(multiplayerModeIds) {
		missingIds := make([]uint64, 0, len(multiplayerModeIds)-len(multiplayerModes))
		for _, id := range multiplayerModeIds {
			found := false
			for _, item := range multiplayerModes {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		multiplayerModes, err = client.MultiplayerModes.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		multiplayerModes = append(multiplayerModes, multiplayerModes...)
		err = SaveItems(endpoint.EPMultiplayerModes, multiplayerModes)
		if err != nil {
			return nil, err
		}
	}
	res.MultiplayerModes = multiplayerModes

	res.Name = game.Name

	if game.ParentGame != nil {
		res.ParentGame = model.GameId(game.ParentGame.Id)
	}

	platformIds := make([]uint64, 0, len(game.Platforms))
	for _, g := range game.Platforms {
		platformIds = append(platformIds, g.Id)
	}
	platforms, err := GetItemsByIds[pb.Platform](endpoint.EPPlatforms, platformIds)
	if err != nil {
		return nil, err
	}
	if len(platforms) != len(platformIds) {
		missingIds := make([]uint64, 0, len(platformIds)-len(platforms))
		for _, id := range platformIds {
			found := false
			for _, item := range platforms {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		platforms, err = client.Platforms.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		platforms = append(platforms, platforms...)
		err = SaveItems(endpoint.EPPlatforms, platforms)
		if err != nil {
			return nil, err
		}
	}
	res.Platforms = platforms

	playerPerspectiveIds := make([]uint64, 0, len(game.PlayerPerspectives))
	for _, g := range game.PlayerPerspectives {
		playerPerspectiveIds = append(playerPerspectiveIds, g.Id)
	}
	playerPerspectives, err := GetItemsByIds[pb.PlayerPerspective](endpoint.EPPlayerPerspectives, playerPerspectiveIds)
	if err != nil {
		return nil, err
	}
	if len(playerPerspectives) != len(playerPerspectiveIds) {
		missingIds := make([]uint64, 0, len(playerPerspectiveIds)-len(playerPerspectives))
		for _, id := range playerPerspectiveIds {
			found := false
			for _, item := range playerPerspectives {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		playerPerspectives, err = client.PlayerPerspectives.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		playerPerspectives = append(playerPerspectives, playerPerspectives...)
		err = SaveItems(endpoint.EPPlayerPerspectives, playerPerspectives)
		if err != nil {
			return nil, err
		}
	}
	res.PlayerPerspectives = playerPerspectives

	res.Rating = game.Rating
	res.RatingCount = game.RatingCount

	releaseDateIds := make([]uint64, 0, len(game.ReleaseDates))
	for _, g := range game.ReleaseDates {
		releaseDateIds = append(releaseDateIds, g.Id)
	}
	releaseDates, err := GetItemsByIds[pb.ReleaseDate](endpoint.EPReleaseDates, releaseDateIds)
	if err != nil {
		return nil, err
	}
	if len(releaseDates) != len(releaseDateIds) {
		missingIds := make([]uint64, 0, len(releaseDateIds)-len(releaseDates))
		for _, id := range releaseDateIds {
			found := false
			for _, item := range releaseDates {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		releaseDates, err = client.ReleaseDates.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		releaseDates = append(releaseDates, releaseDates...)
		err = SaveItems(endpoint.EPReleaseDates, releaseDates)
		if err != nil {
			return nil, err
		}
	}
	res.ReleaseDates = releaseDates

	screenshotIds := make([]uint64, 0, len(game.Screenshots))
	for _, g := range game.Screenshots {
		screenshotIds = append(screenshotIds, g.Id)
	}
	screenshots, err := GetItemsByIds[pb.Screenshot](endpoint.EPScreenshots, screenshotIds)
	if err != nil {
		return nil, err
	}
	if len(screenshots) != len(screenshotIds) {
		missingIds := make([]uint64, 0, len(screenshotIds)-len(screenshots))
		for _, id := range screenshotIds {
			found := false
			for _, item := range screenshots {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		screenshots, err = client.Screenshots.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		screenshots = append(screenshots, screenshots...)
		err = SaveItems(endpoint.EPScreenshots, screenshots)
		if err != nil {
			return nil, err
		}
	}
	res.Screenshots = screenshots

	similarGamesIds := make([]uint64, 0, len(game.SimilarGames))
	for _, g := range game.SimilarGames {
		similarGamesIds = append(similarGamesIds, g.Id)
	}
	res.SimilarGames = similarGamesIds

	res.Slug = game.Slug

	standaloneExpansionsIds := make([]uint64, 0, len(game.StandaloneExpansions))
	for _, g := range game.StandaloneExpansions {
		standaloneExpansionsIds = append(standaloneExpansionsIds, g.Id)
	}
	res.StandaloneExpansions = standaloneExpansionsIds

	res.Storyline = game.Storyline
	res.Summary = game.Summary

	res.Tags = game.Tags

	themeIds := make([]uint64, 0, len(game.Themes))
	for _, g := range game.Themes {
		themeIds = append(themeIds, g.Id)
	}
	themes, err := GetItemsByIds[pb.Theme](endpoint.EPThemes, themeIds)
	if err != nil {
		return nil, err
	}
	if len(themes) != len(themeIds) {
		missingIds := make([]uint64, 0, len(themeIds)-len(themes))
		for _, id := range themeIds {
			found := false
			for _, item := range themes {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		themes, err = client.Themes.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		themes = append(themes, themes...)
		err = SaveItems(endpoint.EPThemes, themes)
		if err != nil {
			return nil, err
		}
	}
	res.Themes = themes

	res.TotalRating = game.TotalRating
	res.TotalRatingCount = game.TotalRatingCount

	res.UpdatedAt = game.UpdatedAt

	res.Url = game.Url

	if game.VersionParent != nil {
		res.VersionParent = model.GameId(game.VersionParent.Id)
	}

	res.VersionTitle = game.VersionTitle

	videoIds := make([]uint64, 0, len(game.Videos))
	for _, g := range game.Videos {
		videoIds = append(videoIds, g.Id)
	}
	videos, err := GetItemsByIds[pb.GameVideo](endpoint.EPGameVideos, videoIds)
	if err != nil {
		return nil, err
	}
	if len(videos) != len(videoIds) {
		missingIds := make([]uint64, 0, len(videoIds)-len(videos))
		for _, id := range videoIds {
			found := false
			for _, item := range videos {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		videos, err = client.GameVideos.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		videos = append(videos, videos...)
		err = SaveItems(endpoint.EPGameVideos, videos)
		if err != nil {
			return nil, err
		}
	}
	res.Videos = videos

	websiteIds := make([]uint64, 0, len(game.Websites))
	for _, g := range game.Websites {
		websiteIds = append(websiteIds, g.Id)
	}
	websites, err := GetItemsByIds[pb.Website](endpoint.EPWebsites, websiteIds)
	if err != nil {
		return nil, err
	}
	if len(websites) != len(websiteIds) {
		missingIds := make([]uint64, 0, len(websiteIds)-len(websites))
		for _, id := range websiteIds {
			found := false
			for _, item := range websites {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		websites, err = client.Websites.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		websites = append(websites, websites...)
		err = SaveItems(endpoint.EPWebsites, websites)
		if err != nil {
			return nil, err
		}
	}
	res.Websites = websites

	remakesIds := make([]uint64, 0, len(game.Remakes))
	for _, g := range game.Remakes {
		remakesIds = append(remakesIds, g.Id)
	}
	res.Remakes = remakesIds

	remastersIds := make([]uint64, 0, len(game.Remasters))
	for _, g := range game.Remasters {
		remastersIds = append(remastersIds, g.Id)
	}
	res.Remasters = remastersIds

	expandedGamesIds := make([]uint64, 0, len(game.ExpandedGames))
	for _, g := range game.ExpandedGames {
		expandedGamesIds = append(expandedGamesIds, g.Id)
	}
	res.ExpandedGames = expandedGamesIds

	portsIds := make([]uint64, 0, len(game.Ports))
	for _, g := range game.Ports {
		portsIds = append(portsIds, g.Id)
	}
	res.Ports = portsIds

	forksIds := make([]uint64, 0, len(game.Forks))
	for _, g := range game.Forks {
		forksIds = append(forksIds, g.Id)
	}
	res.Forks = forksIds

	languageSupportIds := make([]uint64, 0, len(game.LanguageSupports))
	for _, g := range game.LanguageSupports {
		languageSupportIds = append(languageSupportIds, g.Id)
	}
	languageSupports, err := GetItemsByIds[pb.LanguageSupport](endpoint.EPLanguageSupports, languageSupportIds)
	if err != nil {
		return nil, err
	}
	if len(languageSupports) != len(languageSupportIds) {
		missingIds := make([]uint64, 0, len(languageSupportIds)-len(languageSupports))
		for _, id := range languageSupportIds {
			found := false
			for _, item := range languageSupports {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		languageSupports, err = client.LanguageSupports.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		languageSupports = append(languageSupports, languageSupports...)
		err = SaveItems(endpoint.EPLanguageSupports, languageSupports)
		if err != nil {
			return nil, err
		}
	}
	res.LanguageSupports = languageSupports

	gameLocalizationIds := make([]uint64, 0, len(game.GameLocalizations))
	for _, g := range game.GameLocalizations {
		gameLocalizationIds = append(gameLocalizationIds, g.Id)
	}
	gameLocalizations, err := GetItemsByIds[pb.GameLocalization](endpoint.EPGameLocalizations, gameLocalizationIds)
	if err != nil {
		return nil, err
	}
	if len(gameLocalizations) != len(gameLocalizationIds) {
		missingIds := make([]uint64, 0, len(gameLocalizationIds)-len(gameLocalizations))
		for _, id := range gameLocalizationIds {
			found := false
			for _, item := range gameLocalizations {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		gameLocalizations, err = client.GameLocalizations.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		gameLocalizations = append(gameLocalizations, gameLocalizations...)
		err = SaveItems(endpoint.EPGameLocalizations, gameLocalizations)
		if err != nil {
			return nil, err
		}
	}
	res.GameLocalizations = gameLocalizations

	collectionIds := make([]uint64, 0, len(game.Collections))
	for _, g := range game.Collections {
		collectionIds = append(collectionIds, g.Id)
	}
	collections, err := GetItemsByIds[pb.Collection](endpoint.EPCollections, collectionIds)
	if err != nil {
		return nil, err
	}
	if len(collections) != len(collectionIds) {
		missingIds := make([]uint64, 0, len(collectionIds)-len(collections))
		for _, id := range collectionIds {
			found := false
			for _, item := range collections {
				if item.GetId() == id {
					found = true
					break
				}
			}
			if !found {
				missingIds = append(missingIds, id)
			}
		}
		collections, err = client.Collections.GetByIDs(missingIds)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collections...)
		err = SaveItems(endpoint.EPCollections, collections)
		if err != nil {
			return nil, err
		}
	}
	res.Collections = collections

	res.GameStatus = game.GameStatus
	res.GameType = game.GameType

	res.AllNames = make([]string, 0, len(alternativeNames)+1)
	res.AllNames = append(res.AllNames, game.Name)
	for _, item := range alternativeNames {
		res.AllNames = append(res.AllNames, item.Name)
	}

	return res, nil
}

func GetGameById(id uint64) (*model.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var game model.Game
	err := GetInstance().GameCollection.FindOne(ctx, bson.M{"id": id}).Decode(&game)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	return &game, nil
}

func GetAllItemsIDs[T any](e endpoint.Name) ([]uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ids []uint64
	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"id": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	type IdGetter interface {
		GetId() uint64
	}

	for cursor.Next(ctx) {
		var item *T
		err := cursor.Decode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to decode item: %w", err)
		}
		ids = append(ids, any(item).(IdGetter).GetId())
	}

	return ids, nil
}
