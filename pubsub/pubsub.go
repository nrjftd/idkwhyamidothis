package pubsub

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func SubscribeToTaskEvent(RedisClient *redis.Client) {
	pubsub := RedisClient.Subscribe(context.Background(), "task_event")
	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			parts := strings.Split(msg.Payload, "|")
			eventType := parts[0]
			taskID := parts[1]
			switch eventType {
			case "task_created":
				fmt.Println("New task created: ", taskID)
			case "task_deleted", "task_updated":
				fmt.Println("task changed: ", taskID)
				RedisClient.Del(context.Background(), "task:"+taskID)
			}
		}
	}()

}
