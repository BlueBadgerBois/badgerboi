package cache

import(
	"github.com/go-redis/redis"
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

