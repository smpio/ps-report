package datastore

import (
	"time"

	"github.com/go-redis/redis"
)

// Connection is datastore connection
type Connection struct {
	client *redis.Client
}

// New returns new datastore connection
func New(addr string, password string, db int) (*Connection, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	return &Connection{
		client: client,
	}, nil
}

// Write writes to datastore
func (c *Connection) Write(key string, value string, expiration time.Duration) error {
	return c.client.Set(key, value, expiration).Err()
}

// Close closes datastore connection and frees all resources
func (c *Connection) Close() error {
	return c.client.Close()
}
