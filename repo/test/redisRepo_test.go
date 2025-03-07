package test

import (
	"context"
	repo "jwt2/repo/src"
	"log"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	client *redis.Client
	cmd    redis.Cmdable
)

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("Failed to start miniredis: %v", err)
	}
	client = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	cmd = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	code := m.Run()
	os.Exit(code)

}
func TestSet(t *testing.T) {
	ctx := context.Background()
	redisRepo := repo.NewRedisRepo(cmd)
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "existing key",
			key:     "existing key",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "non-existing key",
			key:     "missing",
			value:   "test",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := redisRepo.Set(ctx, tt.key, tt.value, time.Minute)
			if tt.wantErr {
				assert.Nil(t, err)
				return
			}
			assert.NoError(t, err)

			got, err := redisRepo.Get(ctx, tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	redisRepo := repo.NewRedisRepo(cmd)
	err := redisRepo.Set(ctx, "existing key", "this is value", time.Minute)
	require.NoError(t, err)
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "existing key",
			key:     "existing key",
			value:   "this is value",
			wantErr: false,
		},
		{
			name:    "non exist",
			key:     "non exist",
			value:   "value",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := redisRepo.Get(ctx, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", got)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tt.value, got)
		})
	}

}

func TestPubSub(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, cmd, "cmd is nil")
	require.NotNil(t, client, "client is nil")
	redisRepo := repo.NewRedisRepo(cmd)
	channel := "test"
	message := "abc"
	msgCh, err := redisRepo.Sub(ctx, channel)
	require.NoError(t, err)
	go func() {
		time.Sleep(100 * time.Microsecond)
		err := redisRepo.Pub(ctx, channel, message)
		require.NoError(t, err)
	}()
	select {
	case msg := <-msgCh:
		assert.Equal(t, message, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive message")
	}

}
