package db

import (
	"context"
	"fmt"
	"igdb-database/model"
	"time"

	"github.com/bestnite/go-igdb/endpoint"
	pb "github.com/bestnite/go-igdb/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func IsGamesAggregated(games []*pb.Game) (map[uint64]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(games))*200*time.Millisecond)
	defer cancel()

	ids := make([]uint64, 0, len(games))
	for _, game := range games {
		ids = append(ids, game.Id)
	}

	cursor, err := GetInstance().GameCollection.Find(ctx, bson.M{"id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %v", err)
	}

	res := make(map[uint64]bool, len(games))
	g := []*model.Game{}
	err = cursor.All(ctx, &g)
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %v", err)
	}
	for _, game := range g {
		res[game.Id] = true
	}

	return res, nil
}

func SaveGame(game *model.Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if game.MId.IsZero() {
		game.MId = bson.NewObjectID()
	}
	filter := bson.M{"_id": game.MId}
	update := bson.M{"$set": game}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := GetInstance().GameCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func ConvertGame(game *pb.Game) (*model.Game, error) {
	res := &model.Game{}

	res.Id = game.Id

	ageRatings, err := GetItemsByIGDBGameID[pb.AgeRating](endpoint.EPAgeRatings, game.Id)
	if err != nil {
		return nil, err
	}
	res.AgeRatings = make([]*pb.AgeRating, 0, len(ageRatings))
	for _, item := range ageRatings {
		res.AgeRatings = append(res.AgeRatings, item.Item)
	}

	res.AggregatedRating = game.AggregatedRating
	res.AggregatedRatingCount = game.AggregatedRatingCount

	alternativeNames, err := GetItemsByIGDBGameID[pb.AlternativeName](endpoint.EPAlternativeNames, game.Id)
	if err != nil {
		return nil, err
	}
	res.AlternativeNames = make([]*pb.AlternativeName, 0, len(alternativeNames))
	for _, item := range alternativeNames {
		res.AlternativeNames = append(res.AlternativeNames, item.Item)
	}

	Artworks, err := GetItemsByIGDBGameID[pb.Artwork](endpoint.EPArtworks, game.Id)
	if err != nil {
		return nil, err
	}
	res.Artworks = make([]*pb.Artwork, 0, len(Artworks))
	for _, item := range Artworks {
		res.Artworks = append(res.Artworks, item.Item)
	}

	bundlesIds := make([]uint64, 0, len(game.Bundles))
	for _, g := range game.Bundles {
		bundlesIds = append(bundlesIds, g.Id)
	}
	res.Bundles = bundlesIds

	covers, err := GetItemsByIGDBGameID[pb.Cover](endpoint.EPCovers, game.Id)
	if err != nil {
		return nil, err
	}
	if len(covers) != 0 {
		res.Cover = covers[0].Item
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

	externalGames, err := GetItemsByIGDBGameID[pb.ExternalGame](endpoint.EPExternalGames, game.Id)
	if err != nil {
		return nil, err
	}
	res.ExternalGames = make([]*pb.ExternalGame, 0, len(externalGames))
	for _, item := range externalGames {
		res.ExternalGames = append(res.ExternalGames, item.Item)
	}

	res.FirstReleaseDate = game.FirstReleaseDate

	res.Franchise = nil

	franchises, err := GetItemsByIGDBGameID[pb.Franchise](endpoint.EPFranchises, game.Id)
	if err != nil {
		return nil, err
	}
	res.Franchises = make([]*pb.Franchise, 0, len(franchises))
	for _, item := range franchises {
		res.Franchises = append(res.Franchises, item.Item)
	}

	gameEngines, err := GetItemsByIGDBGameID[pb.GameEngine](endpoint.EPGameEngines, game.Id)
	if err != nil {
		return nil, err
	}
	res.GameEngines = make([]*pb.GameEngine, 0, len(gameEngines))
	for _, item := range gameEngines {
		res.GameEngines = append(res.GameEngines, item.Item)
	}

	gameModes, err := GetItemsByIGDBGameID[pb.GameMode](endpoint.EPGameModes, game.Id)
	if err != nil {
		return nil, err
	}
	res.GameModes = make([]*pb.GameMode, 0, len(gameModes))
	for _, item := range gameModes {
		res.GameModes = append(res.GameModes, item.Item)
	}

	genres, err := GetItemsByIGDBGameID[pb.Genre](endpoint.EPGenres, game.Id)
	if err != nil {
		return nil, err
	}
	res.Genres = make([]*pb.Genre, 0, len(genres))
	for _, item := range genres {
		res.Genres = append(res.Genres, item.Item)
	}

	res.Hypes = game.Hypes

	involvedCompanies, err := GetItemsByIGDBGameID[pb.InvolvedCompany](endpoint.EPInvolvedCompanies, game.Id)
	if err != nil {
		return nil, err
	}
	res.InvolvedCompanies = make([]*pb.InvolvedCompany, 0, len(involvedCompanies))
	for _, item := range involvedCompanies {
		res.InvolvedCompanies = append(res.InvolvedCompanies, item.Item)
	}

	keywords, err := GetItemsByIGDBGameID[pb.Keyword](endpoint.EPKeywords, game.Id)
	if err != nil {
		return nil, err
	}
	res.Keywords = make([]*pb.Keyword, 0, len(keywords))
	for _, item := range keywords {
		res.Keywords = append(res.Keywords, item.Item)
	}

	multiplayerModes, err := GetItemsByIGDBGameID[pb.MultiplayerMode](endpoint.EPMultiplayerModes, game.Id)
	if err != nil {
		return nil, err
	}
	res.MultiplayerModes = make([]*pb.MultiplayerMode, 0, len(multiplayerModes))
	for _, item := range multiplayerModes {
		res.MultiplayerModes = append(res.MultiplayerModes, item.Item)
	}

	res.Name = game.Name

	if game.ParentGame != nil {
		res.ParentGame = model.GameId(game.ParentGame.Id)
	}

	platforms, err := GetItemsByIGDBGameID[pb.Platform](endpoint.EPPlatforms, game.Id)
	if err != nil {
		return nil, err
	}
	res.Platforms = make([]*pb.Platform, 0, len(platforms))
	for _, item := range platforms {
		res.Platforms = append(res.Platforms, item.Item)
	}

	playerPerspectives, err := GetItemsByIGDBGameID[pb.PlayerPerspective](endpoint.EPPlayerPerspectives, game.Id)
	if err != nil {
		return nil, err
	}
	res.PlayerPerspectives = make([]*pb.PlayerPerspective, 0, len(playerPerspectives))
	for _, item := range playerPerspectives {
		res.PlayerPerspectives = append(res.PlayerPerspectives, item.Item)
	}

	res.Rating = game.Rating
	res.RatingCount = game.RatingCount

	releaseDates, err := GetItemsByIGDBGameID[pb.ReleaseDate](endpoint.EPReleaseDates, game.Id)
	if err != nil {
		return nil, err
	}
	res.ReleaseDates = make([]*pb.ReleaseDate, 0, len(releaseDates))
	for _, item := range releaseDates {
		res.ReleaseDates = append(res.ReleaseDates, item.Item)
	}

	screenshots, err := GetItemsByIGDBGameID[pb.Screenshot](endpoint.EPScreenshots, game.Id)
	if err != nil {
		return nil, err
	}
	res.Screenshots = make([]*pb.Screenshot, 0, len(screenshots))
	for _, item := range screenshots {
		res.Screenshots = append(res.Screenshots, item.Item)
	}

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

	themes, err := GetItemsByIGDBGameID[pb.Theme](endpoint.EPThemes, game.Id)
	if err != nil {
		return nil, err
	}
	res.Themes = make([]*pb.Theme, 0, len(themes))
	for _, item := range themes {
		res.Themes = append(res.Themes, item.Item)
	}

	res.TotalRating = game.TotalRating
	res.TotalRatingCount = game.TotalRatingCount

	res.UpdatedAt = game.UpdatedAt

	res.Url = game.Url

	if game.VersionParent != nil {
		res.VersionParent = model.GameId(game.VersionParent.Id)
	}

	res.VersionTitle = game.VersionTitle

	videos, err := GetItemsByIGDBGameID[pb.GameVideo](endpoint.EPGameVideos, game.Id)
	if err != nil {
		return nil, err
	}
	res.Videos = make([]*pb.GameVideo, 0, len(videos))
	for _, item := range videos {
		res.Videos = append(res.Videos, item.Item)
	}

	websites, err := GetItemsByIGDBGameID[pb.Website](endpoint.EPWebsites, game.Id)
	if err != nil {
		return nil, err
	}
	res.Websites = make([]*pb.Website, 0, len(websites))
	for _, item := range websites {
		res.Websites = append(res.Websites, item.Item)
	}

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

	languageSupports, err := GetItemsByIGDBGameID[pb.LanguageSupport](endpoint.EPLanguageSupports, game.Id)
	if err != nil {
		return nil, err
	}
	res.LanguageSupports = make([]*pb.LanguageSupport, 0, len(languageSupports))
	for _, item := range languageSupports {
		res.LanguageSupports = append(res.LanguageSupports, item.Item)
	}

	gameLocalizations, err := GetItemsByIGDBGameID[pb.GameLocalization](endpoint.EPGameLocalizations, game.Id)
	if err != nil {
		return nil, err
	}
	res.GameLocalizations = make([]*pb.GameLocalization, 0, len(gameLocalizations))
	for _, item := range gameLocalizations {
		res.GameLocalizations = append(res.GameLocalizations, item.Item)
	}

	collections, err := GetItemsByIGDBGameID[pb.Collection](endpoint.EPCollections, game.Id)
	if err != nil {
		return nil, err
	}
	res.Collections = make([]*pb.Collection, 0, len(collections))
	for _, item := range collections {
		res.Collections = append(res.Collections, item.Item)
	}

	res.GameStatus = nil
	res.GameType = nil

	res.AllNames = make([]string, 0, len(alternativeNames)+1)
	res.AllNames = append(res.AllNames, game.Name)
	for _, item := range alternativeNames {
		res.AllNames = append(res.AllNames, item.Item.Name)
	}

	return res, nil
}

func GetGameByIGDBID(id uint64) (*model.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var game model.Game
	err := GetInstance().GameCollection.FindOne(ctx, bson.M{"id": id}).Decode(&game)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %v", err)
	}
	return &game, nil
}
