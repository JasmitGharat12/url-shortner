// package db

// import (
// 	"context"
// 	"fmt"
// 	"github.com/go-redis/redis/v8"
// 	"hash/fnv"
// 	"net/url"
// 	"strings"
// )

// var ctx = context.Background()

// type RedisStore struct {
// 	Client *redis.Client
// }

// func NewRedisStore() *RedisStore {
// 	rdb := redis.NewClient(&redis.Options{
		
// 		Addr: "localhost:6379",
// 		DB:   0,
// 	})
// 	return &RedisStore{Client: rdb}
// }

// func NewTestRedisStore() *RedisStore {
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr: "localhost:6379",
// 		DB:   1,
// 	})
// 	return &RedisStore{Client: rdb}
// }

// // SaveURL stores the original URL and generates a short URL if not exists
// func (s *RedisStore) SaveURL(originalURL string) (string, error) {
// 	shortURL, err := s.Client.Get(ctx, originalURL).Result()
// 	if err == redis.Nil {
// 		// Generate new short URL
// 		shortURL = s.generateShortURL(originalURL)

// 		// Save original -> short mapping
// 		if err = s.Client.Set(ctx, originalURL, shortURL, 0).Err(); err != nil {
// 			return "", err
// 		}

// 		// Save short -> original mapping
// 		if err = s.Client.Set(ctx, shortURL, originalURL, 0).Err(); err != nil {
// 			return "", err
// 		}

// 		// Update domain count
// 		domain, err := s.getDomain(originalURL)
// 		if err != nil {
// 			return "", err
// 		}
// 		if err = s.Client.Incr(ctx, fmt.Sprintf("domain:%s", domain)).Err(); err != nil {
// 			return "", err
// 		}
// 	} else if err != nil {
// 		return "", err
// 	}
// 	return shortURL, nil
// }

// // GetOriginalURL retrieves the original URL from Redis using the short URL
// func (s *RedisStore) GetOriginalURL(shortURL string) (string, error) {
// 	originalURL, err := s.Client.Get(ctx, shortURL).Result()
// 	if err == redis.Nil {
// 		return "", fmt.Errorf("URL not found")
// 	} else if err != nil {
// 		return "", err
// 	}
// 	return originalURL, nil
// }

// // GetDomainCounts retrieves the counts of shortened URLs per domain
// func (s *RedisStore) GetDomainCounts() (map[string]int, error) {
// 	keys, err := s.Client.Keys(ctx, "domain:*").Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	domainCounts := make(map[string]int)
// 	for _, key := range keys {
// 		count, err := s.Client.Get(ctx, key).Int()
// 		if err != nil {
// 			return nil, err
// 		}
// 		domain := strings.TrimPrefix(key, "domain:")
// 		domainCounts[domain] = count
// 	}
// 	return domainCounts, nil
// }

// func (s *RedisStore) getDomain(originalURL string) (string, error) {
// 	parsedURL, err := url.Parse(originalURL)
// 	if err != nil {
// 		return "", err
// 	}
// 	return strings.TrimPrefix(parsedURL.Host, "www."), nil
// }

// // returns a short URL.
// func (s *RedisStore) generateShortURL(originalURL string) string {
// 	h := fnv.New32a()
// 	h.Write([]byte(originalURL))
// 	return fmt.Sprintf("%x", h.Sum32())
// }




package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"hash/fnv"
	"net/url"
	"os"
	"strings"
)

var ctx = context.Background()

type RedisStore struct {
	Client *redis.Client
}

func NewRedisStore() *RedisStore {
	// Read from environment variables (works both in Docker & local)
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	// Optional: Ping check to fail fast if Redis is unreachable
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("⚠️ Failed to connect to Redis at %s: %v\n", addr, err)
	}

	return &RedisStore{Client: rdb}
}

func NewTestRedisStore() *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	return &RedisStore{Client: rdb}
}

// SaveURL stores the original URL and generates a short URL if not exists
func (s *RedisStore) SaveURL(originalURL string) (string, error) {
	shortURL, err := s.Client.Get(ctx, originalURL).Result()
	if err == redis.Nil {
		// Generate new short URL
		shortURL = s.generateShortURL(originalURL)

		// Save original -> short mapping
		if err = s.Client.Set(ctx, originalURL, shortURL, 0).Err(); err != nil {
			return "", err
		}

		// Save short -> original mapping
		if err = s.Client.Set(ctx, shortURL, originalURL, 0).Err(); err != nil {
			return "", err
		}

		// Update domain count
		domain, err := s.getDomain(originalURL)
		if err != nil {
			return "", err
		}
		if err = s.Client.Incr(ctx, fmt.Sprintf("domain:%s", domain)).Err(); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return shortURL, nil
}

// GetOriginalURL retrieves the original URL from Redis using the short URL
func (s *RedisStore) GetOriginalURL(shortURL string) (string, error) {
	originalURL, err := s.Client.Get(ctx, shortURL).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("URL not found")
	} else if err != nil {
		return "", err
	}
	return originalURL, nil
}

// GetDomainCounts retrieves the counts of shortened URLs per domain
func (s *RedisStore) GetDomainCounts() (map[string]int, error) {
	keys, err := s.Client.Keys(ctx, "domain:*").Result()
	if err != nil {
		return nil, err
	}

	domainCounts := make(map[string]int)
	for _, key := range keys {
		count, err := s.Client.Get(ctx, key).Int()
		if err != nil {
			return nil, err
		}
		domain := strings.TrimPrefix(key, "domain:")
		domainCounts[domain] = count
	}
	return domainCounts, nil
}

func (s *RedisStore) getDomain(originalURL string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(parsedURL.Host, "www."), nil
}

// returns a short URL.
func (s *RedisStore) generateShortURL(originalURL string) string {
	h := fnv.New32a()
	h.Write([]byte(originalURL))
	return fmt.Sprintf("%x", h.Sum32())
}
