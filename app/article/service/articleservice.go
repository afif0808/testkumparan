package articleservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/afif0808/testkumparan/app/article"
	"github.com/afif0808/testkumparan/app/cache"
	"github.com/afif0808/testkumparan/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func cacheKeyword(parameters map[string]interface{}) string {
	k := "articles"
	if parameters != nil {
		if parameters["author"] != nil {
			k += ":author:" + fmt.Sprint(parameters["author"])
		}
		if parameters["s"] != nil {
			k += ":search:" + fmt.Sprint(parameters["s"])
		}
	}
	return k
}

func cacheByIDKeyword(id primitive.ObjectID) string {
	return "articles:single:" + id.String()
}

func cacheToArticle(cache interface{}) (*models.Article, error) {
	bytes, ok := cache.([]byte)
	if !ok {
		return nil, errors.New("the cache is not in *models.Article format")
	}
	art := &models.Article{}
	err := json.Unmarshal(bytes, &art)
	if err != nil {
		return nil, err
	}
	return art, nil
}
func cacheToArticleArray(cache interface{}) ([]models.Article, error) {
	bytes, ok := cache.([]byte)
	if !ok {
		return nil, errors.New("the cache is not in *models.Article format")
	}
	arts := []models.Article{}
	err := json.Unmarshal(bytes, &arts)
	if err != nil {
		return nil, err
	}
	return arts, nil
}

func GetMany(getFromDB article.GetManyFunc, getCache cache.GetFunc, storeCache cache.StoreFunc) article.GetManyFunc {
	return func(ctx context.Context, parameters map[string]interface{}) ([]models.Article, error) {
		cache, err := getCache(ctx, cacheKeyword(parameters))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if cache != nil {
			arts, err := cacheToArticleArray(cache)
			if err != nil {
				// in this case i decided the system won't to stop the code and return error
				// instead the process continue with fetching the new one from database.
				// But still logs the problem
				log.Println(err)
			} else if len(arts) > 0 {
				return arts, nil
			}

		}
		arts, err := getFromDB(ctx, parameters)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if len(arts) > 0 {
			go func() {
				bytes, err := json.Marshal(arts)
				if err != nil {
					log.Println(err)
				}
				err = storeCache(ctx, cacheKeyword(parameters), bytes)
				if err != nil {
					log.Println(err)
				}
			}()
		}
		return arts, nil
	}
}

func GetByID(getFromDB article.GetByIDFunc, getCache cache.GetFunc, storeCache cache.StoreFunc) article.GetByIDFunc {
	return func(ctx context.Context, id primitive.ObjectID) (*models.Article, error) {
		cache, err := getCache(ctx, cacheByIDKeyword(id))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if cache != nil {
			art, err := cacheToArticle(cache)
			if err != nil {
				// in this case i decided the system won't to stop the code and return error
				// instead the process continue with fetching the new one from database.
				// But still logs the problem
				log.Println(err)
			} else if art != nil {

				return art, nil
			}
		}
		art, err := getFromDB(ctx, id)
		if err != nil {
			return nil, err
		}
		if art != nil {
			go func() {
				bytes, err := json.Marshal(art)
				if err != nil {
					log.Println(err)
				}
				err = storeCache(ctx, cacheByIDKeyword(id), bytes)
				if err != nil {

					log.Println(err)
				}
			}()
		}
		return art, nil
	}
}

func Create(
	create article.CreateFunc,
	storeCache cache.StoreFunc,
	deleteCacheByKeyPattern cache.DeleteFunc,
	updateCache article.UpdateCacheFunc,
) article.CreateFunc {
	return func(ctx context.Context, art models.Article) (models.Article, error) {
		art, err := create(ctx, art)
		if err != nil {
			log.Println(err)
		}
		go func() {
			updateCache(ctx, &art, cacheByIDKeyword(art.ID))
			updateCache(ctx, &art, "articles")
			updateCache(ctx, &art, "articles:author:"+art.Author)

			// i decided to delete those keys ( for this test ) due to complexity( of code ).
			err = deleteCacheByKeyPattern(ctx, "search")
			if err != nil {
				log.Println(err)
			}

		}()

		return art, nil
	}
}

func UpdateCache(getCache cache.GetFunc, storeCache cache.StoreFunc) article.UpdateCacheFunc {
	return func(ctx context.Context, newArt *models.Article, key string) error {

		if strings.Contains(key, "single") {
			bytes, err := json.Marshal(newArt)
			if err != nil {
				log.Println(err)
			}
			err = storeCache(ctx, key, bytes)
			if err != nil {
				log.Println(err)
			}
			return err
		}

		arts := []models.Article{}
		cache, err := getCache(ctx, key)
		if err != nil {
			log.Println(err)
			return err
		}

		if cache != nil {
			arts, err = cacheToArticleArray(cache)
			if err != nil {
				log.Println(err, key)
				return err
			}
		}
		arts = append(arts, *newArt)
		bytes, err := json.Marshal(arts)
		if err != nil {
			log.Println(err)
			return err
		}
		err = storeCache(ctx, key, bytes)
		return err
	}
}
