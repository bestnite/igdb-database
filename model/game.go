package model

import (
	pb "github.com/bestnite/go-igdb/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GameIds []uint64
type GameId uint64

type Game struct {
	MId                   bson.ObjectID           `bson:"_id,omitempty" json:"_id,omitempty"`
	Id                    uint64                  `json:"id,omitempty"`
	AgeRatings            []*pb.AgeRating         `json:"age_ratings,omitempty"`
	AggregatedRating      float64                 `json:"aggregated_rating,omitempty"`
	AggregatedRatingCount int32                   `json:"aggregated_rating_count,omitempty"`
	AlternativeNames      []*pb.AlternativeName   `json:"alternative_names,omitempty"`
	Artworks              []*pb.Artwork           `json:"artworks,omitempty"`
	Bundles               GameIds                 `json:"bundles,omitempty"`
	Cover                 *pb.Cover               `json:"cover,omitempty"`
	CreatedAt             *timestamppb.Timestamp  `json:"created_at,omitempty"`
	Dlcs                  GameIds                 `json:"dlcs,omitempty"`
	Expansions            GameIds                 `json:"expansions,omitempty"`
	ExternalGames         []*pb.ExternalGame      `json:"external_games,omitempty"`
	FirstReleaseDate      *timestamppb.Timestamp  `json:"first_release_date,omitempty"`
	Franchise             *pb.Franchise           `json:"franchise,omitempty"`
	Franchises            []*pb.Franchise         `json:"franchises,omitempty"`
	GameEngines           []*pb.GameEngine        `json:"game_engines,omitempty"`
	GameModes             []*pb.GameMode          `json:"game_modes,omitempty"`
	Genres                []*pb.Genre             `json:"genres,omitempty"`
	Hypes                 int32                   `json:"hypes,omitempty"`
	InvolvedCompanies     []*pb.InvolvedCompany   `json:"involved_companies,omitempty"`
	Keywords              []*pb.Keyword           `json:"keywords,omitempty"`
	MultiplayerModes      []*pb.MultiplayerMode   `json:"multiplayer_modes,omitempty"`
	Name                  string                  `json:"name,omitempty"`
	ParentGame            GameId                  `json:"parent_game,omitempty"`
	Platforms             []*pb.Platform          `json:"platforms,omitempty"`
	PlayerPerspectives    []*pb.PlayerPerspective `json:"player_perspectives,omitempty"`
	Rating                float64                 `json:"rating,omitempty"`
	RatingCount           int32                   `json:"rating_count,omitempty"`
	ReleaseDates          []*pb.ReleaseDate       `json:"release_dates,omitempty"`
	Screenshots           []*pb.Screenshot        `json:"screenshots,omitempty"`
	SimilarGames          GameIds                 `json:"similar_games,omitempty"`
	Slug                  string                  `json:"slug,omitempty"`
	StandaloneExpansions  GameIds                 `json:"standalone_expansions,omitempty"`
	Storyline             string                  `json:"storyline,omitempty"`
	Summary               string                  `json:"summary,omitempty"`
	Tags                  []int32                 `json:"tags,omitempty"`
	Themes                []*pb.Theme             `json:"themes,omitempty"`
	TotalRating           float64                 `json:"total_rating,omitempty"`
	TotalRatingCount      int32                   `json:"total_rating_count,omitempty"`
	UpdatedAt             *timestamppb.Timestamp  `json:"updated_at,omitempty"`
	Url                   string                  `json:"url,omitempty"`
	VersionParent         GameId                  `json:"version_parent,omitempty"`
	VersionTitle          string                  `json:"version_title,omitempty"`
	Videos                []*pb.GameVideo         `json:"videos,omitempty"`
	Websites              []*pb.Website           `json:"websites,omitempty"`
	Remakes               GameIds                 `json:"remakes,omitempty"`
	Remasters             GameIds                 `json:"remasters,omitempty"`
	ExpandedGames         GameIds                 `json:"expanded_games,omitempty"`
	Ports                 GameIds                 `json:"ports,omitempty"`
	Forks                 GameIds                 `json:"forks,omitempty"`
	LanguageSupports      []*pb.LanguageSupport   `json:"language_supports,omitempty"`
	GameLocalizations     []*pb.GameLocalization  `json:"game_localizations,omitempty"`
	Collections           []*pb.Collection        `json:"collections,omitempty"`
	GameStatus            *pb.GameStatus          `json:"game_status,omitempty"`
	GameType              *pb.GameType            `json:"game_type,omitempty"`

	AllNames []string `bson:"all_names,omitempty"`
}
