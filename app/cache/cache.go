package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

func Get(redisPool *redis.Pool) GetFunc {
	return func(ctx context.Context, keyword string) (interface{}, error) {
		conn := redisPool.Get()
		defer conn.Close()
		reply, err := conn.Do("GET", keyword)
		if err != nil {
			return nil, err
		}
		if reply == nil {
			return nil, nil
		}
		return reply, nil
	}
}

func Store(redisPool *redis.Pool) StoreFunc {
	return func(ctx context.Context, keyword string, data interface{}) error {
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("SET", keyword, data)
		if err != nil {
			return err
		}
		return nil
	}
}

func ClearAllCache(redisPool *redis.Pool) ClearAllCacheFunc {
	return func(ctx context.Context) error {
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("flushall")
		if err != nil {
			log.Println(err)
		}
		return err
	}
}

func GetKeys(redisPool *redis.Pool) GetKeysFunc {
	return func(ctx context.Context, pattern string) ([]string, error) {
		conn := redisPool.Get()
		keys := []string{}
		reply, err := conn.Do("KEYS", "*"+pattern+"*")
		if err != nil {
			return nil, err
		}
		if tmp, ok := reply.([]interface{}); ok {
			for _, v := range tmp {
				keys = append(keys, fmt.Sprintf("%s", v))
			}
		} else {
			return nil, fmt.Errorf("error encountered upon fetching cache keys")
		}
		return keys, nil
	}
}
func DeleteByKey(redisPool *redis.Pool) DeleteFunc {
	return func(ctx context.Context, keyword string) error {
		conn := redisPool.Get()
		_, err := conn.Do("DEL", keyword)
		if err != nil {
			log.Println(err)
		}
		return err
	}
}
func DeleteByKeyPattern(redisPool *redis.Pool) DeleteFunc {
	return func(ctx context.Context, keyword string) error {
		keys, err := GetKeys(redisPool)(ctx, keyword)
		if err != nil {
			return err
		}
		deleteCache := DeleteByKey(redisPool)
		for _, k := range keys {
			err = deleteCache(ctx, k)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
