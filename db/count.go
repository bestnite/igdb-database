package db

import (
	"context"
	"fmt"
	"time"

	"github.com/bestnite/go-igdb/endpoint"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Count struct {
	Theme uint64 `json:"theme,omitempty"`
	Genre uint64 `json:"genre,omitempty"`
	Count int64  `json:"count,omitempty"`
}

func GetCountByThemeId(themeId uint64) (*Count, error) {
	var count Count
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := GetInstance().CountCollection.FindOne(ctx, bson.M{"theme": themeId}).Decode(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}
	return &count, nil
}

func GetCountByGenreId(genreId uint64) (*Count, error) {
	var count Count
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := GetInstance().CountCollection.FindOne(ctx, bson.M{"genre": genreId}).Decode(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}
	return &count, nil
}

func SaveCount(count *Count) error {
	var filter bson.M
	if count.Genre != 0 {
		filter = bson.M{"genre": count.Genre}
	}
	if count.Theme != 0 {
		filter = bson.M{"theme": count.Theme}
	}
	update := bson.M{"$set": count}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().CountCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func AddThemeCount(themeId uint64) error {
	filter := bson.M{"theme": themeId}
	update := bson.M{"$inc": bson.M{"count": 1}}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().CountCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func AddGenreCount(genreId uint64) error {
	filter := bson.M{"genre": genreId}
	update := bson.M{"$inc": bson.M{"count": 1}}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().CountCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func MinusThemeCount(themeId uint64) error {
	filter := bson.M{"theme": themeId}
	update := bson.M{"$inc": bson.M{"count": -1}}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().CountCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func MinusGenreCount(genreId uint64) error {
	filter := bson.M{"genre": genreId}
	update := bson.M{"$inc": bson.M{"count": -1}}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().CountCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func CountTheme(themeId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	filter := bson.M{"themes.id": themeId}
	count, err := GetInstance().Collections[endpoint.EPGames].CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count theme: %w", err)
	}
	return count, nil
}

func CountGenre(genreId uint64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	filter := bson.M{"genres.id": genreId}
	count, err := GetInstance().Collections[endpoint.EPGames].CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count genre: %w", err)
	}
	return count, nil
}
