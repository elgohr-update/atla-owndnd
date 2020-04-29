package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/atla/owndnd/pkg/db"
	"github.com/atla/owndnd/pkg/entities"
	e "github.com/atla/owndnd/pkg/entities"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenericRepo ...
type GenericRepo struct {
	db         *db.Client
	collection string
	generator  func() interface{}
}

type entityGenerator func() interface{}
type elementCollector func(element interface{})

// FindByID ...
func (repo *GenericRepo) FindByID(id string) (interface{}, error) {

	result := repo.db.FindByID(repo.collection, id)

	if result != nil {
		entity := repo.generator()
		if err := result.Decode(entity); err != nil {
			log.WithField("Error", err).Error("Error decoding entity")
			return nil, errors.New("entity not found")
		}
		return entity, nil
	}
	return nil, errors.New("entity not found")
}

// FindAllWithParam returns all entities
func (repo *GenericRepo) FindAllWithParam(params *db.QueryParams, collector elementCollector) error {

	cursor, err := repo.db.FindAllWithParams(repo.collection, params)

	if err != nil {
		log.WithField("collection", repo.collection).WithField("cursor", cursor).Fatal(err)
	}

	for cursor.Next(context.TODO()) {
		elem := repo.generator()
		err := cursor.Decode(elem)
		if err != nil {
			return errors.New("Could not decode element")
		}
		collector(elem)
	}
	return nil
}

// FindAll returns all entities
func (repo *GenericRepo) FindAll(collector elementCollector) error {

	//var results []interface{}
	cursor, err := repo.db.FindAll(repo.collection)

	if err != nil {
		log.WithField("collection", repo.collection).WithField("cursor", cursor).Fatal(err)
	}

	for cursor.Next(context.TODO()) {
		elem := repo.generator()
		err := cursor.Decode(elem)
		if err != nil {
			return errors.New("Could not decode element")
		}
		collector(elem)
	}
	return nil
}

// Store stores a new entity
func (repo *GenericRepo) Store(entity interface{}) (interface{}, error) {

	// update creation time
	//item.Created = time.Now()

	entity.(e.Entity).ID = primitive.NewObjectID()

	fmt.Println("ID: " + entity.(e.Entity).ID)

	if result, error := repo.db.InsertOne(repo.collection, entity); error != nil {
		log.WithField("Error", error).Error("error during insertion")
		return nil, error
	} else {
		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {

			if n, ok := entity.(*entities.Entity); ok {
				n.ID = entities.EntityID(oid)
			}
		}
	}

	return entity, nil

}

// Update an existing entity
func (repo *GenericRepo) Update(item interface{}, id e.EntityID) error {

	if result, error := repo.db.UpdateOneByID(repo.collection, primitive.ObjectID(id), item); error != nil {
		log.WithField("Error", error).Error("error during insertion")
		return error
	} else {
		log.WithField("Generic Update", result).Info("updated entity")
	}

	return nil
}
