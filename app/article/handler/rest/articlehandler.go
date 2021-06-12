package articleresthandler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/afif0808/gomongohelper"
	"github.com/afif0808/testkumparan/app/article"
	articlemongorepo "github.com/afif0808/testkumparan/app/article/repository/mongo"
	articleservice "github.com/afif0808/testkumparan/app/article/service"
	"github.com/afif0808/testkumparan/app/cache"
	"github.com/afif0808/testkumparan/app/models"
	"github.com/afif0808/testkumparan/app/worker"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func setMongoDBIndexes(db *mongo.Database) {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "title", Value: "text"}, {Key: "body", Value: "text"}},
		},
		{
			Keys: bson.D{{Key: "author", Value: 1}},
		},
	}
	res, err := db.Collection("articles").Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(res)
}

func parameters(ectx echo.Context) map[string]interface{} {
	urlValues := ectx.QueryParams()
	parameters := map[string]interface{}{}

	for k, v := range urlValues {
		parameters[k] = strings.Join(v[:], ",")
	}
	return parameters
}

func authorFilter(filter bson.M, parameters map[string]interface{}) (bson.M, error) {
	if parameters != nil && parameters["author"] != nil {
		filter["author"] = parameters["author"]
	}
	return filter, nil
}

func InjectArticleRESTHandler(ee *echo.Echo, db *mongo.Database, redisPool *redis.Pool, workerDispatcher *worker.Dispatcher) {
	setMongoDBIndexes(db)

	storeCache := cache.Store(redisPool)
	getCache := cache.Get(redisPool)
	deleteCacheByKeyPartern := cache.DeleteByKeyPattern(redisPool)

	// i prefer mongo db FTS since it's fast and i think it fits the need of this test
	// But there is a disadvantage. mongo db full text search does not support partial match yet
	// So not gonna lie , i had doubt on choosing between full text and regex search.
	//
	fullTextSearch := gomongohelper.FullTextSearch("s")
	// regexSearch := gomongohelper.RegexSearch([]string{"title", "body"}, "s")

	getManyRepo := articlemongorepo.GetMany(db, fullTextSearch, authorFilter)

	getManyService := articleservice.GetMany(getManyRepo, getCache, storeCache)
	createRepo := articlemongorepo.Create(db)
	updateArticleCache := articleservice.UpdateCache(getCache, storeCache)
	createService := articleservice.Create(createRepo, storeCache, deleteCacheByKeyPartern, updateArticleCache)
	getByIDRepo := articlemongorepo.GetByID(db)
	getByIDService := articleservice.GetByID(getByIDRepo, getCache, storeCache)

	ee.GET("/articles", GetMany(getManyService, workerDispatcher))
	ee.GET("/articles/:id", GetByID(getByIDService, workerDispatcher))
	ee.POST("/articles", Create(createService))

}

func GetMany(getMany article.GetManyFunc, workerDispatcher *worker.Dispatcher) echo.HandlerFunc {
	return func(ectx echo.Context) error {
		p := parameters(ectx)
		ctx := ectx.Request().Context()
		arts := []models.Article{}
		done := make(chan bool)
		var err error
		workerDispatcher.JobQueue <- func() {
			var tmp []models.Article
			tmp, err = getMany(ctx, p)
			arts = tmp
			done <- true
		}
		<-done
		// arts, err := getMany(ctx, nil)

		if err != nil {
			return ectx.JSON(http.StatusInternalServerError, err)
		}
		return ectx.JSON(http.StatusOK, arts)
	}
}

func GetByID(get article.GetByIDFunc, workerDispatcher *worker.Dispatcher) echo.HandlerFunc {
	return func(ectx echo.Context) error {
		ctx := ectx.Request().Context()
		id, err := primitive.ObjectIDFromHex(ectx.Param("id"))
		if err != nil {
			return ectx.JSON(http.StatusBadRequest, err)
		}
		done := make(chan bool)
		var art *models.Article
		workerDispatcher.JobQueue <- func() {
			art, err = get(ctx, id)
			done <- true
		}
		<-done
		if err != nil {
			return ectx.JSON(http.StatusInternalServerError, err)
		}
		return ectx.JSON(http.StatusOK, art)
	}
}

func Create(create article.CreateFunc) echo.HandlerFunc {
	return func(ectx echo.Context) error {
		ctx := ectx.Request().Context()
		art := models.Article{}
		err := ectx.Bind(&art)
		if err != nil {
			return ectx.JSON(http.StatusBadRequest, err)
		}
		art, err = create(ctx, art)
		if err != nil {
			return ectx.JSON(http.StatusInternalServerError, err)
		}
		return ectx.JSON(http.StatusCreated, art)
	}
}
