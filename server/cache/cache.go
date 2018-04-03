package cache

import(
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"time"
)

type Cache struct {
	Client *redis.Client
}

func redisUrl() string {
	url := os.Getenv("REDIS_URL")
	return url
}

func (cache *Cache) Init() {
	client := redis.NewClient(&redis.Options{
		Addr: redisUrl(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cache.Client = client	

	_, err := cache.Client.Ping().Result()
	if err != nil {
		fmt.Println("**************************************")
		fmt.Println("!!!!! Failed connecting to redis !!!!!")
		fmt.Println("**************************************")
	} else {
		fmt.Println("Connected to redis")
	}
}

func (cache *Cache) SetKeyWithExpirationInSecs(key string, val string, expSecs uint) error {
	secondsDelta := time.Duration(expSecs) * time.Second
	err := cache.Client.Set(key, val, secondsDelta).Err()

	if err != nil {
		return errors.New("Could not set key " + key + "with expiration")
	}

	return nil
}

func (cache *Cache) GetKeyWithStringVal(key string) (string, error) {
	val, err := cache.Client.Get(key).Result()
	return val, err
}
