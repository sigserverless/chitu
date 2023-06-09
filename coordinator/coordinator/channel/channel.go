package channel

import (
	"context"
	"coordinator/object"
	"encoding/json"

	redis "github.com/go-redis/redis/v8"
)

var GlobalChannel *Channel

type Channel struct {
	RedisClient *redis.Client
}

func (ch *Channel) NewObject(object *object.Object, key string) error {

	b, err := json.Marshal(object)
	if err != nil {
		return err
	}
	if _, err2 := ch.RedisClient.Set(context.Background(), key, string(b), -1).Result(); err2 != nil {
		return err2
	}
	return nil
}

func (ch *Channel) SetObject(object *object.Object, key string) error {
	b, err := json.Marshal(object)
	if err != nil {
		return err
	}
	if _, err2 := ch.RedisClient.Set(context.Background(), key, string(b), -1).Result(); err2 != nil {
		return err2
	}
	return nil
}

func (ch *Channel) GetObject(key string) (object *object.Object, err error) {
	var f string
	if f, err = ch.RedisClient.Get(context.Background(), key).Result(); err != nil {
		return
	}
	err = json.Unmarshal([]byte(f), &object)
	if err != nil {
		return
	}
	return
}

func (ch *Channel) Clear() {
	ch.RedisClient.FlushDB(context.Background())
}
