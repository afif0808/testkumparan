package main

import (
	"context"
	"log"
	"os"
	"time"

	articleresthandler "github.com/afif0808/testkumparan/app/article/handler/rest"
	"github.com/afif0808/testkumparan/app/worker"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func redisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   50,
		MaxActive: 10000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379")
			// Connection error handling
			if err != nil {
				log.Printf("ERROR: fail initializing the redis pool: %s", err.Error())
				os.Exit(1)
			}
			return conn, err
		},
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rp := redisPool()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	db := client.Database("test-kumparan")

	ee := echo.New()
	ee.Pre(func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(ectx echo.Context) error {
			ctx, cancel := context.WithTimeout(ectx.Request().Context(), time.Second*10)
			defer cancel()
			ectx.SetRequest(ectx.Request().WithContext(ctx))

			return hf(ectx)
		}
	})
	workerDispatcher := worker.NewDispatcher(1000)
	workerDispatcher.Run()
	articleresthandler.InjectArticleRESTHandler(ee, db, rp, workerDispatcher)
	ee.Start(":484")

}
