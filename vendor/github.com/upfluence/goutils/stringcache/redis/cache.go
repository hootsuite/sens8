package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/upfluence/goutils/stringcache"
)

const redisTimeout = 20 * time.Second

type Cache struct {
	conn redis.Conn
	key  string
}

func NewCache(uri, key string) (*Cache, error) {
	conn, err := redis.DialURL(
		uri,
		redis.DialConnectTimeout(redisTimeout),
		redis.DialReadTimeout(redisTimeout),
		redis.DialWriteTimeout(redisTimeout),
	)

	if err != nil {
		return nil, err
	}

	return &Cache{conn: conn, key: key}, nil
}

func (c *Cache) Has(k string) (bool, error) {
	return redis.Bool(c.conn.Do("SISMEMBER", c.key, k))
}

func (c *Cache) Add(k string) error {
	_, err := c.conn.Do("SADD", c.key, k)

	return err
}

func (c *Cache) Delete(k string) error {
	e, err := redis.Bool(c.conn.Do("SISMEMBER", c.key, k))

	if err != nil {
		return err
	}

	if !e {
		return stringcache.ErrNotFound
	}

	_, err = c.conn.Do("SREM", c.key, k)

	return err
}
