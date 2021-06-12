package articlemongorepo

import (
	"context"
	"log"
	"time"

	"github.com/afif0808/gomongohelper"
	"github.com/afif0808/testkumparan/app/article"
	"github.com/afif0808/testkumparan/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "articles"

func GetMany(db *mongo.Database, filters ...gomongohelper.FilterFunc) article.GetManyFunc {
	return func(ctx context.Context, parameters map[string]interface{}) ([]models.Article, error) {
		filter := bson.M{}

		if parameters != nil {
			for _, f := range filters {
				var err error
				filter, err = f(filter, parameters)
				if err != nil {
					return nil, err
				}
			}
		}

		fo := options.Find()
		// i prefer to sort by id instead of created_at as id is indexed
		// and object id is also created based on timestamp
		fo.SetSort(bson.M{"_id": -1})

		cursor, err := db.Collection(collectionName).Find(ctx, filter, fo)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		arts := []models.Article{}
		err = cursor.All(ctx, &arts)
		if err != nil {
			return nil, err
		}

		return arts, nil
	}
}

func GetByID(db *mongo.Database) article.GetByIDFunc {
	return func(ctx context.Context, id primitive.ObjectID) (*models.Article, error) {
		parameters := map[string]interface{}{"id": id}
		arts, err := GetMany(db, gomongohelper.IDFilter())(ctx, parameters)
		if err != nil {
			return nil, err
		}
		if len(arts) < 1 {
			return nil, nil
		}
		return &arts[0], nil
	}
}

func Create(db *mongo.Database) article.CreateFunc {
	return func(ctx context.Context, art models.Article) (models.Article, error) {
		art.ID = primitive.NewObjectID()
		art.CreatedAt = time.Now()
		_, err := db.Collection(collectionName).InsertOne(ctx, art)
		if err != nil {
			return art, err
		}
		return art, nil
	}
}
