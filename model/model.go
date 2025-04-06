package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Item[T any] struct {
	Item *T            `bson:"item"`
	MId  bson.ObjectID `bson:"_id"`
}

func NewItem[T any](item *T) *Item[T] {
	return &Item[T]{
		Item: item,
	}
}

func NewItems[T any](items []*T) []*Item[T] {
	var result []*Item[T]
	for _, item := range items {
		result = append(result, NewItem(item))
	}
	return result
}
