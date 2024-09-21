package services

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCacheService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := NewMockCacheService(ctrl)

	// Test Set method
	t.Run("Set", func(t *testing.T) {
		ctx := context.Background()
		key := "test_key"
		value := "test_value"

		mockCache.EXPECT().Set(ctx, key, value).Return(nil)

		err := mockCache.Set(ctx, key, value)
		assert.NoError(t, err)
	})

	// Test Get method
	t.Run("Get", func(t *testing.T) {
		ctx := context.Background()
		key := "test_key"
		var value string

		mockCache.EXPECT().Get(ctx, key, gomock.Any()).SetArg(2, "test_value").Return(nil)

		err := mockCache.Get(ctx, key, &value)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	// Test Delete method
	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()
		key := "test_key"

		mockCache.EXPECT().Delete(ctx, key).Return(nil)

		err := mockCache.Delete(ctx, key)
		assert.NoError(t, err)
	})

	// Test HSet method
	t.Run("HSet", func(t *testing.T) {
		ctx := context.Background()
		key := "hash_key"
		field := "field1"
		value := "value1"

		mockCache.EXPECT().HSet(ctx, key, field, value).Return(nil)

		err := mockCache.HSet(ctx, key, field, value)
		assert.NoError(t, err)
	})

	// Test HGet method
	t.Run("HGet", func(t *testing.T) {
		ctx := context.Background()
		key := "hash_key"
		field := "field1"
		expectedValue := "value1"
		var actualValue string

		mockCache.EXPECT().HGet(ctx, key, field).Return(expectedValue, nil)
		
		// Call the method under test and assign the result to actualValue
		actualValue, err := mockCache.HGet(ctx, key, field)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if actualValue != expectedValue {
			t.Errorf("expected %s, got %s", expectedValue, actualValue)
		}
	})

	// Test Pipeline method
	t.Run("Pipeline", func(t *testing.T) {
		ctx := context.Background()
		pipelineFunc := func(pipe redis.Pipeliner) error {
			// Dummy pipeline operations
			return nil
		}

		mockCache.EXPECT().Pipeline(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(redis.Pipeliner) error) error {
			return fn(nil) // Passing nil for simplicity
		})

		err := mockCache.Pipeline(ctx, pipelineFunc)
		assert.NoError(t, err)
	})

	// Test SAdd method
	t.Run("SAdd", func(t *testing.T) {
		ctx := context.Background()
		key := "set_key"
		members := []interface{}{"member1", "member2"}

		mockCache.EXPECT().SAdd(ctx, key, members...).Return(nil)

		err := mockCache.SAdd(ctx, key, members...)
		assert.NoError(t, err)
	})

	// Test SMembers method
	t.Run("SMembers", func(t *testing.T) {
		ctx := context.Background()
		key := "set_key"
		expectedMembers := []string{"member1", "member2"}

		mockCache.EXPECT().SMembers(ctx, key).Return(expectedMembers, nil)

		members, err := mockCache.SMembers(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, expectedMembers, members)
	})

	// Test ZAdd method
	t.Run("ZAdd", func(t *testing.T) {
		ctx := context.Background()
		key := "zset_key"
		members := []*redis.Z{
			{Score: 1.0, Member: "member1"},
			{Score: 2.0, Member: "member2"},
		}

		// Convert members from []*redis.Z to []interface{}
		args := make([]interface{}, len(members))
		for i, member := range members {
			args[i] = member
		}

		// Set up expectation with converted arguments
		mockCache.EXPECT().ZAdd(ctx, key, args...).Return(nil)

		// Call the method under test
		err := mockCache.ZAdd(ctx, key, members...)
		
		// Add your assertions here
		assert.NoError(t, err)
	})

	// Test ZRange method
	t.Run("ZRange", func(t *testing.T) {
		ctx := context.Background()
		key := "zset_key"
		start := int64(0)
		stop := int64(-1)
		expectedRange := []string{"member1", "member2"}

		mockCache.EXPECT().ZRange(ctx, key, start, stop).Return(expectedRange, nil)

		members, err := mockCache.ZRange(ctx, key, start, stop)
		assert.NoError(t, err)
		assert.Equal(t, expectedRange, members)
	})

	// Test Incr method
	t.Run("Incr", func(t *testing.T) {
		ctx := context.Background()
		key := "counter"
		expectedValue := int64(1)

		mockCache.EXPECT().Incr(ctx, key).Return(expectedValue, nil)

		value, err := mockCache.Incr(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	})

	// Test Decr method
	t.Run("Decr", func(t *testing.T) {
		ctx := context.Background()
		key := "counter"
		expectedValue := int64(-1)

		mockCache.EXPECT().Decr(ctx, key).Return(expectedValue, nil)

		value, err := mockCache.Decr(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	})

	// Test Error Cases
	t.Run("Set_Error", func(t *testing.T) {
		ctx := context.Background()
		key := "test_key"
		value := "test_value"
		expectedErr := errors.New("set error")

		mockCache.EXPECT().Set(ctx, key, value).Return(expectedErr)

		err := mockCache.Set(ctx, key, value)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("Get_Error", func(t *testing.T) {
		ctx := context.Background()
		key := "test_key"
		var value string
		expectedErr := errors.New("get error")

		mockCache.EXPECT().Get(ctx, key, gomock.Any()).Return(expectedErr)

		err := mockCache.Get(ctx, key, &value)
		assert.ErrorIs(t, err, expectedErr)
	})

	// Similarly, you can add error case tests for other methods as needed
}