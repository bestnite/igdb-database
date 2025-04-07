package db

import (
	"context"
	"fmt"
	"igdb-database/config"
	"igdb-database/model"
	"log"
	"sync"
	"time"

	"github.com/bestnite/go-igdb/endpoint"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	once     sync.Once
	instance *MongoDB
)

type MongoDB struct {
	client         *mongo.Client
	Collections    map[endpoint.EndpointName]*mongo.Collection
	GameCollection *mongo.Collection
}

func GetInstance() *MongoDB {
	once.Do(func() {
		clientOptions := options.Client().ApplyURI(fmt.Sprintf(
			"mongodb://%s:%s@%s:%v",
			config.C().Database.User,
			config.C().Database.Password,
			config.C().Database.Host,
			config.C().Database.Port,
		)).SetConnectTimeout(3 * time.Second)

		client, err := mongo.Connect(clientOptions)
		if err != nil {
			log.Fatalf("failed to connect to mongodb: %v", err)
		}
		instance = &MongoDB{
			client:      client,
			Collections: make(map[endpoint.EndpointName]*mongo.Collection),
		}

		for _, e := range endpoint.AllEndpoints {
			instance.Collections[e] = client.Database(config.C().Database.Database).Collection(string(e))
		}

		instance.GameCollection = client.Database(config.C().Database.Database).Collection("game_details")
		instance.createIndex()
	})

	return instance
}

func (m *MongoDB) createIndex() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*60*time.Second)
	defer cancel()

	textIndexMap := map[endpoint.EndpointName]string{
		endpoint.EPGames:            "item.name",
		endpoint.EPAlternativeNames: "item.name",
	}

	for e, idx := range textIndexMap {
		_, err := m.Collections[e].Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: idx, Value: "text"},
			},
		})
		if err != nil {
			log.Printf("failed to create index %s for %s: %v", idx, string(e), err)
		}
	}

	indexMap := map[endpoint.EndpointName][]string{
		endpoint.EPAlternativeNames:         {"item.game.id"},
		endpoint.EPArtworks:                 {"item.game.id"},
		endpoint.EPCollectionMemberships:    {"item.game.id"},
		endpoint.EPCovers:                   {"item.game.id"},
		endpoint.EPExternalGames:            {"item.game.id"},
		endpoint.EPGameEngines:              {"item.game.id"},
		endpoint.EPGameLocalizations:        {"item.game.id"},
		endpoint.EPGameVersions:             {"item.game.id"},
		endpoint.EPGameVersionFeatureValues: {"item.game.id"},
		endpoint.EPGameVideos:               {"item.game.id"},
		endpoint.EPInvolvedCompanies:        {"item.game.id"},
		endpoint.EPLanguageSupports:         {"item.game.id"},
		endpoint.EPMultiplayerModes:         {"item.game.id"},
		endpoint.EPReleaseDates:             {"item.game.id"},
		endpoint.EPScreenshots:              {"item.game.id"},
		endpoint.EPWebsites:                 {"item.game.id"},
		endpoint.EPGames:                    {"item.parent_game.id", "item.version_parent.id"},
	}

	for e, idxes := range indexMap {
		for _, idx := range idxes {
			_, err := m.Collections[e].Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{
					{Key: idx, Value: 1},
				},
			})
			if err != nil {
				log.Printf("failed to create index %s for %s: %v", idx, string(e), err)
			}
		}
	}

	for _, e := range endpoint.AllEndpoints {
		if e == endpoint.EPWebhooks || e == endpoint.EPSearch || e == endpoint.EPPopularityPrimitives {
			continue
		}
		_, err := m.Collections[e].Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "item.id", Value: 1},
			},
		})
		if err != nil {
			log.Printf("failed to create index item.id for %s: %v", string(e), err)
		}
	}
}

func SaveItem[T any](e endpoint.EndpointName, item *model.Item[T]) error {
	if item.MId.IsZero() {
		item.MId = bson.NewObjectID()
	}
	filter := bson.M{"_id": item.MId}
	update := bson.M{"$set": item}
	opts := options.UpdateOne().SetUpsert(true)

	coll := GetInstance().Collections[e]
	if coll == nil {
		return fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save item to %s: %w", string(e), err)
	}
	return nil
}

func SaveItems[T any](e endpoint.EndpointName, items []*model.Item[T]) error {
	var models []mongo.WriteModel

	for _, item := range items {
		if item.MId.IsZero() {
			item.MId = bson.NewObjectID()
		}
		filter := bson.M{"_id": item.MId}
		update := bson.M{"$set": item}
		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, model)
	}

	coll := GetInstance().Collections[e]
	if coll == nil {
		return fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(items))*200*time.Millisecond)
	defer cancel()
	_, err := coll.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("failed to save items in bulk %s: %w", string(e), err)
	}
	return nil
}

func CountItems(e endpoint.EndpointName) (int64, error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return 0, fmt.Errorf("collection not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count items %s: %w", string(e), err)
	}
	return count, nil
}

func GetItemByIGDBID[T any](e endpoint.EndpointName, id uint64) (*model.Item[T], error) {
	var item model.Item[T]
	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := coll.FindOne(ctx, bson.M{"item.id": id}).Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("failed to get item %s: %w", string(e), err)
	}
	return &item, nil
}

func GetItemsByIGDBIDs[T any](e endpoint.EndpointName, ids []uint64) (map[uint64]*model.Item[T], error) {
	if len(ids) == 0 {
		return nil, nil
	}

	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := coll.Find(ctx, bson.M{"item.id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("failed to get items %s: %w", string(e), err)
	}

	type IdGetter interface {
		GetId() uint64
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(len(ids))*200*time.Millisecond)
	defer cancel()
	res := make(map[uint64]*model.Item[T])
	for cursor.Next(ctx) {
		item := model.Item[T]{}
		err := cursor.Decode(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to decode item %s: %w", string(e), err)
		}
		if v, ok := any(item.Item).(IdGetter); ok {
			res[v.GetId()] = &item
		} else {
			return nil, fmt.Errorf("failed to get id from item %s: %w", string(e), err)
		}
	}

	return res, nil
}

func RemoveItemByID(e endpoint.EndpointName, id bson.ObjectID) error {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to remove game: %w", err)
	}
	return nil
}

func RemoveItemsByID(e endpoint.EndpointName, ids []bson.ObjectID) error {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return fmt.Errorf("failed to remove games: %w", err)
	}
	return nil
}

func RemoveDuplicateItems(e endpoint.EndpointName) error {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return fmt.Errorf("collection not found")
	}
	pipeline := bson.A{
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$item.id"},
				{Key: "docs", Value: bson.D{
					{Key: "$push", Value: "$_id"},
				}},
			}},
		},
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "$expr", Value: bson.D{
					{Key: "$gt", Value: bson.A{
						bson.D{{Key: "$size", Value: "$docs"}},
						1,
					}},
				}},
			}},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate: %w", err)
	}
	var results []struct {
		ID   uint64          `bson:"_id"`
		Docs []bson.ObjectID `bson:"docs"`
	}
	err = cursor.All(ctx, &results)
	if err != nil {
		return fmt.Errorf("failed to get results: %w", err)
	}

	removedIds := make([]bson.ObjectID, 0, len(results))

	for _, result := range results {
		removedIds = append(removedIds, result.Docs[1:]...)
	}

	err = RemoveItemsByID(e, removedIds)
	if err != nil {
		return fmt.Errorf("failed to remove duplicate games: %w", err)
	}

	return nil
}

func GetItemsByIGDBGameID[T any](e endpoint.EndpointName, id uint64) ([]*model.Item[T], error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"item.game.id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to get items %s: %w", string(e), err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var items []*model.Item[T]
	err = cursor.All(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to decode items %s: %w", string(e), err)
	}

	return items, nil
}

func GetItemsPagnated[T any](e endpoint.EndpointName, offset int64, limit int64) ([]*model.Item[T], error) {

	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(limit)*200*time.Millisecond)
	defer cancel()
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSkip(offset).SetLimit(limit).SetSort(bson.D{{Key: "item.id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to get items %s: %w", string(e), err)
	}

	var items []*model.Item[T]
	err = cursor.All(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to decode items %s: %w", string(e), err)
	}

	return items, nil
}
