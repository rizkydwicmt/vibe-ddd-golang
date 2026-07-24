package redis

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"vibe-ddd-golang/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PaginationResult struct {
	CurrentPage int   `json:"currentPage"`
	PerPage     int   `json:"perPage"`
	TotalItems  int64 `json:"totalItems"`
	TotalPages  int   `json:"totalPages"`
	Data        any   `json:"data"`
}

// GetPageOffset implements offset-based pagination (simple but less efficient for large datasets)
func (r *Client) GetPageOffset(key string, page, pageSize int) (*PaginationResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	offset := int64((page - 1) * pageSize)
	limit := offset + int64(pageSize) - 1

	// Get items from sorted set
	items, err := r.Client.ZRange(r.ctx, key, offset, limit).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get page items: %w", err)
	}

	// Parse JSON data
	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(item), &parsed); err != nil {
			data = append(data, map[string]any{"value": item})
		} else {
			data = append(data, parsed)
		}
	}

	// Get total count
	total, err := r.Client.ZCard(r.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &PaginationResult{
		CurrentPage: page,
		PerPage:     pageSize,
		TotalItems:  total,
		TotalPages:  totalPages,
		Data:        data,
	}, nil
}

// AddData adds new data items to the sorted set with automatic scoring
func (r *Client) AddData(key string, data []map[string]any, ttl *time.Duration) error {
	if len(data) == 0 {
		return fmt.Errorf("no data provided")
	}

	pipe := r.Client.Pipeline()

	for i, item := range data {
		if item["id"] == nil {
			item["id"] = uuid.New()
		}
		// Marshal the data to JSON
		jsonData, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item %d: %w", i, err)
		}

		// Use timestamp + index as score for chronological ordering
		// You can customize this scoring logic based on your needs
		score := float64(time.Now().UnixNano()) + float64(i)

		// If the item has a custom score field, use it
		if scoreVal, exists := item["score"]; exists {
			if s, ok := scoreVal.(float64); ok {
				score = s
			} else if s, ok := scoreVal.(int); ok {
				score = float64(s)
			}
		}

		pipe.ZAdd(r.ctx, key, redis.Z{
			Score:  score,
			Member: string(jsonData),
		})

		if ttl != nil {
			pipe.Expire(r.ctx, key, *ttl)
		}
	}

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to add data: %w", err)
	}

	return nil
}

// AppendData performs upsert operation - updates existing items or adds new ones to increase data in key
func (r *Client) AppendData(key string, data []map[string]any, ttl *time.Duration) error {
	if len(data) == 0 {
		return fmt.Errorf("no data provided")
	}

	// Check if key exists and has data
	keyExists, err := r.Client.ZCard(r.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check if key exists: %w", err)
	}

	// If key doesn't exist or is empty, just use AddData for better performance
	if keyExists == 0 {
		fmt.Printf("Key '%s' is empty, using AddData for bulk insert\n", key)
		return r.AddData(key, data, ttl)
	}

	// Get all existing items to check for duplicates
	existingItems, err := r.Client.ZRange(r.ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get existing items: %w", err)
	}

	// Create a map of existing items by ID for fast lookup
	existingByID := make(map[string]string)    // id -> full JSON item
	existingScores := make(map[string]float64) // id -> score

	for _, existingItem := range existingItems {
		var existing map[string]any
		if err := json.Unmarshal([]byte(existingItem), &existing); err != nil {
			continue // Skip non-JSON items
		}

		if id, exists := existing["id"]; exists {
			idStr := fmt.Sprintf("%v", id)
			existingByID[idStr] = existingItem

			// Get the score of existing item
			scores, err := r.Client.ZMScore(r.ctx, key, existingItem).Result()
			if err == nil && len(scores) > 0 {
				existingScores[idStr] = scores[0]
			}
		}
	}

	pipe := r.Client.Pipeline()
	updatedCount := 0
	addedCount := 0

	for i, item := range data {
		idVal, exists := item["id"]
		if !exists {
			if item["id"] == nil {
				item["id"] = uuid.New()
			}
			idVal = item["id"]
		}

		id := fmt.Sprintf("%v", idVal)

		// Marshal the new/updated data to JSON
		jsonData, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item %d: %w", i, err)
		}

		// Determine score
		var score float64
		if existingScore, exists := existingScores[id]; exists {
			// Item exists - use existing score unless new score is provided
			score = existingScore
			if scoreVal, hasCustomScore := item["score"]; hasCustomScore {
				if s, ok := scoreVal.(float64); ok {
					score = s
				} else if s, ok := scoreVal.(int); ok {
					score = float64(s)
				}
			}

			// Remove old item first
			if oldItem, exists := existingByID[id]; exists {
				pipe.ZRem(r.ctx, key, oldItem)
			}
			updatedCount++
		} else {
			// New item - generate new score
			score = float64(time.Now().UnixNano()) + float64(i)
			if scoreVal, hasCustomScore := item["score"]; hasCustomScore {
				if s, ok := scoreVal.(float64); ok {
					score = s
				} else if s, ok := scoreVal.(int); ok {
					score = float64(s)
				}
			}
			addedCount++
		}

		// Add the item (new or updated)
		pipe.ZAdd(r.ctx, key, redis.Z{
			Score:  score,
			Member: string(jsonData),
		})

		if ttl != nil {
			pipe.Expire(r.ctx, key, *ttl)
		}
	}

	_, err = pipe.Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to assign data: %w", err)
	}

	logger.Debug.Printf("AssignData completed: %d updated, %d added to %s\n", updatedCount, addedCount, key)
	return nil
}

// UpdateData updates existing data items based on an identifier field
func (r *Client) UpdateData(key string, data []map[string]any) error {
	if len(data) == 0 {
		return fmt.Errorf("no data provided")
	}

	for _, item := range data {
		// Each item must have an "id" field for identification
		idVal, exists := item["id"]
		if !exists {
			if item["id"] == nil {
				item["id"] = uuid.New()
			}
			idVal = item["id"]
		}

		id := fmt.Sprintf("%v", idVal)

		// Find existing items with the same ID
		existingItems, err := r.Client.ZRange(r.ctx, key, 0, -1).Result()
		if err != nil {
			return fmt.Errorf("failed to get existing items: %w", err)
		}

		var found bool
		var oldScore float64

		// Search for the item with matching ID
		for _, existingItem := range existingItems {
			var existing map[string]any
			if err := json.Unmarshal([]byte(existingItem), &existing); err != nil {
				continue // Skip non-JSON items
			}

			if existingID := fmt.Sprintf("%v", existing["id"]); existingID == id {
				// Get the score of the existing item
				scores, err := r.Client.ZMScore(r.ctx, key, existingItem).Result()
				if err != nil {
					return fmt.Errorf("failed to get item score: %w", err)
				}
				oldScore = scores[0]

				// Remove the old item
				if err := r.Client.ZRem(r.ctx, key, existingItem).Err(); err != nil {
					return fmt.Errorf("failed to remove old item: %w", err)
				}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("item with id '%s' not found", id)
		}

		// Marshal the updated data
		jsonData, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal updated item: %w", err)
		}

		// Use the old score unless a new score is provided
		score := oldScore
		if scoreVal, exists := item["score"]; exists {
			if s, ok := scoreVal.(float64); ok {
				score = s
			} else if s, ok := scoreVal.(int); ok {
				score = float64(s)
			}
		}

		// Add the updated item
		if err := r.Client.ZAdd(r.ctx, key, redis.Z{
			Score:  score,
			Member: string(jsonData),
		}).Err(); err != nil {
			return fmt.Errorf("failed to add updated item: %w", err)
		}
	}

	return nil
}

// RemoveDataByIDs removes data items by their ID fields (best practice: atomic operations)
func (r *Client) RemoveDataByIDs(key string, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	// Get all items to find matches
	existingItems, err := r.Client.ZRange(r.ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get existing items: %w", err)
	}

	var itemsToRemove []string
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	// Find items that match the IDs
	for _, existingItem := range existingItems {
		var existing map[string]any
		if err := json.Unmarshal([]byte(existingItem), &existing); err != nil {
			continue // Skip non-JSON items
		}

		if existingID := fmt.Sprintf("%v", existing["id"]); idSet[existingID] {
			itemsToRemove = append(itemsToRemove, existingItem)
			delete(idSet, existingID) // Mark as found
		}
	}

	// Check if all IDs were found
	var notFound []string
	for id := range idSet {
		notFound = append(notFound, id)
	}
	if len(notFound) > 0 {
		return fmt.Errorf("items not found with IDs: %v", notFound)
	}

	// Remove all matched items in a single operation
	if len(itemsToRemove) > 0 {
		args := make([]any, len(itemsToRemove))
		for i, item := range itemsToRemove {
			args[i] = item
		}

		if err := r.Client.ZRem(r.ctx, key, args...).Err(); err != nil {
			return fmt.Errorf("failed to remove items: %w", err)
		}
	}

	return nil
}

// RemoveDataByScore removes data items within a score range (best practice for bulk operations)
func (r *Client) RemoveDataByScore(key string, minScore, maxScore float64) (int64, error) {
	minScoreStr := strconv.FormatFloat(minScore, 'f', -1, 64)
	maxScoreStr := strconv.FormatFloat(maxScore, 'f', -1, 64)

	removedCount, err := r.Client.ZRemRangeByScore(r.ctx, key, minScoreStr, maxScoreStr).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to remove items by score: %w", err)
	}

	return removedCount, nil
}

// RemoveDataByRank removes data items by rank/position (best practice for cleanup operations)
func (r *Client) RemoveDataByRank(key string, start, stop int64) (int64, error) {
	// Count how many items will be removed
	count, err := r.Client.ZRange(r.ctx, key, start, stop).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count items to remove: %w", err)
	}

	// Remove items by rank
	if err := r.Client.ZRemRangeByRank(r.ctx, key, start, stop).Err(); err != nil {
		return 0, fmt.Errorf("failed to remove items by rank: %w", err)
	}

	return int64(len(count)), nil
}

// GetById retrieves a specific item by its ID from the sorted set
func (r *Client) GetById(key string, id string) (map[string]any, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Get all items from sorted set
	existingItems, err := r.Client.ZRange(r.ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get existing items: %w", err)
	}

	// Search for the item with matching ID
	for _, existingItem := range existingItems {
		var existing map[string]any
		if err := json.Unmarshal([]byte(existingItem), &existing); err != nil {
			continue // Skip non-JSON items
		}

		if existingID := fmt.Sprintf("%v", existing["id"]); existingID == id {
			return existing, nil
		}
	}

	return nil, fmt.Errorf("item with id '%s' not found", id)
}

// GetByScore retrieves all items within a specified score range from the sorted set
func (r *Client) GetByScore(key string, minScore, maxScore float64) ([]map[string]any, error) {
	minScoreStr := strconv.FormatFloat(minScore, 'f', -1, 64)
	maxScoreStr := strconv.FormatFloat(maxScore, 'f', -1, 64)

	// Get items from sorted set by score range
	items, err := r.Client.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min: minScoreStr,
		Max: maxScoreStr,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get items by score: %w", err)
	}

	// Parse JSON data
	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(item), &parsed); err != nil {
			// If it's not valid JSON, wrap it in a map
			data = append(data, map[string]any{"value": item})
		} else {
			data = append(data, parsed)
		}
	}

	return data, nil
}

// GetAll retrieves all items from the sorted set without pagination
func (r *Client) GetAll(key string) (*PaginationResult, error) {
	items, err := r.Client.ZRange(r.ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all items: %w", err)
	}

	// Parse JSON data
	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(item), &parsed); err != nil {
			data = append(data, map[string]any{"value": item})
		} else {
			data = append(data, parsed)
		}
	}

	// Get total count
	total := int64(len(items))

	return &PaginationResult{
		CurrentPage: 1,
		PerPage:     int(total),
		TotalItems:  total,
		TotalPages:  1,
		Data:        data,
	}, nil
}
