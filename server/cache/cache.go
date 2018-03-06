package cache

import(
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type Cache struct {
	Client *redis.Client
}

func (cache *Cache) Init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cache.Client = client	
}

func (cache *Cache) GetCurrUnixTimeInSecs() (time.Time, error) {
	val, err :=	cache.Client.Time().Result()
	return val, err
}

func (cache *Cache) ExpireKeyAtUnixSeconds(key string, tm time.Time) error {
	_, err :=	cache.Client.ExpireAt(key, tm).Result()
	if err != nil {
		return errors.New("Could not expire key " + key)
	}
	
	return nil
}

func (cache *Cache) ExpireKeyInFuture(key string, numSeconds uint) error {
	currTimeInUnixSecs, err := cache.GetCurrUnixTimeInSecs()
	if err != nil {
		return err
	}

	fmt.Println("current time: ", currTimeInUnixSecs.String())
	secondsDelta := time.Duration(numSeconds) * time.Second
	expireTimeInUnixSecs := currTimeInUnixSecs.Add(secondsDelta)
	fmt.Println("expire time: ", expireTimeInUnixSecs.String())


	err = cache.ExpireKeyAtUnixSeconds(key, expireTimeInUnixSecs)

	return err
}

