package db

import (
	"context"
	"fmt"
	"igdb-database/config"
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
	Collections    map[endpoint.Name]*mongo.Collection
	GameCollection *mongo.Collection
}

func GetInstance() *MongoDB {
	once.Do(func() {
		bsonOpts := &options.BSONOptions{
			UseJSONStructTags: true,
		}

		clientOptions := options.Client().ApplyURI(fmt.Sprintf(
			"mongodb://%s:%s@%s:%v",
			config.C().Database.User,
			config.C().Database.Password,
			config.C().Database.Host,
			config.C().Database.Port,
		)).SetConnectTimeout(3 * time.Second).SetBSONOptions(bsonOpts)

		client, err := mongo.Connect(clientOptions)
		if err != nil {
			log.Fatalf("failed to connect to mongodb: %v", err)
		}
		instance = &MongoDB{
			client:      client,
			Collections: make(map[endpoint.Name]*mongo.Collection),
		}

		for _, e := range endpoint.AllNames {
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

	textIndexMap := map[endpoint.Name]string{
		endpoint.EPGames:            "name",
		endpoint.EPAlternativeNames: "name",
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

	indexMap := map[endpoint.Name][]string{
		endpoint.EPAlternativeNames:         {"game.id"},
		endpoint.EPArtworks:                 {"game.id"},
		endpoint.EPCollectionMemberships:    {"game.id"},
		endpoint.EPCovers:                   {"game.id"},
		endpoint.EPExternalGames:            {"game.id"},
		endpoint.EPGameEngines:              {"game.id"},
		endpoint.EPGameLocalizations:        {"game.id"},
		endpoint.EPGameVersions:             {"game.id"},
		endpoint.EPGameVersionFeatureValues: {"game.id"},
		endpoint.EPGameVideos:               {"game.id"},
		endpoint.EPInvolvedCompanies:        {"game.id"},
		endpoint.EPLanguageSupports:         {"game.id"},
		endpoint.EPMultiplayerModes:         {"game.id"},
		endpoint.EPReleaseDates:             {"game.id"},
		endpoint.EPScreenshots:              {"game.id"},
		endpoint.EPWebsites:                 {"game.id"},
		endpoint.EPGames:                    {"parent_game.id", "version_parent.id"},
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

	for _, e := range endpoint.AllNames {
		if e == endpoint.EPWebhooks || e == endpoint.EPSearch {
			continue
		}
		_, err := m.Collections[e].Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			log.Printf("failed to create index id for %s: %v", string(e), err)
		}
	}

	_, err := m.GameCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("failed to create index id for game_details: %v", err)
	}
}

func CountDocuments(e endpoint.Name) (int64, error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return 0, fmt.Errorf("collection not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count  %s: %w", string(e), err)
	}
	return count, nil
}

func EstimatedDocumentCount(e endpoint.Name) (int64, error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return 0, fmt.Errorf("collection not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count  %s: %w", string(e), err)
	}
	return count, nil
}

func RemoveByID(e endpoint.Name, ids []bson.ObjectID) error {
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

func GetItemsByIds[T any](e endpoint.Name, ids []uint64) ([]*T, error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second+time.Duration(len(ids)*200)*time.Millisecond)
	defer cancel()
	cursor, err := coll.Find(ctx, bson.M{"id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	var items []*T
	err = cursor.All(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	return items, nil
}

func GetItemById[T any](e endpoint.Name, id uint64) (*T, error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var item *T
	err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	return item, nil
}

func SaveItem[T any](e endpoint.Name, item *T) error {
	type IdGetter interface {
		GetId() uint64
	}
	id := any(item).(IdGetter).GetId()
	filter := bson.M{"id": id}
	update := bson.M{"$set": item}
	opts := options.UpdateOne().SetUpsert(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := GetInstance().Collections[e].UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func SaveItems[T any](e endpoint.Name, items []*T) error {
	type IdGetter interface {
		GetId() uint64
	}

	updateModel := make([]mongo.WriteModel, 0, len(items))
	for _, item := range items {
		updateModel = append(updateModel, mongo.NewUpdateOneModel().SetFilter(bson.M{"id": any(item).(IdGetter).GetId()}).SetUpdate(bson.M{"$set": item}).SetUpsert(true))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(items))*200*time.Millisecond)
	defer cancel()
	_, err := GetInstance().Collections[e].BulkWrite(ctx, updateModel)
	if err != nil {
		return err
	}
	return nil
}

func GetItemsPaginated[T any](e endpoint.Name, skip int64, limit int64) ([]*T, error) {
	coll := GetInstance().Collections[e]
	if coll == nil {
		return nil, fmt.Errorf("collection not found")
	}

	opts := options.Find().SetSort(bson.M{"id": 1}).SetSkip(skip).SetLimit(limit)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(limit)*time.Millisecond*200)
	defer cancel()
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get items %s: %w", string(e), err)
	}

	var items []*T
	err = cursor.All(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to get items %s: %w", string(e), err)
	}
	return items, nil
}
