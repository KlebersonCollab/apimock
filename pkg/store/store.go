package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"apimock/pkg/models"
)

// CollectionStore manages thread-safe stateful resource collections
type CollectionStore struct {
	mu          sync.RWMutex
	collections map[string]*models.Collection
	persistPath string
}

// NewCollectionStore creates a new collection store instance
func NewCollectionStore(persistPath string) *CollectionStore {
	store := &CollectionStore{
		collections: make(map[string]*models.Collection),
		persistPath: persistPath,
	}

	if persistPath != "" {
		_ = store.LoadFromFile(persistPath)
	}

	return store
}

// GetCollection returns a collection by ID or name
func (cs *CollectionStore) GetCollection(nameOrID string) (*models.Collection, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	key := strings.ToLower(strings.TrimSpace(nameOrID))
	// Try direct name match
	if col, exists := cs.collections[key]; exists {
		return col, true
	}
	// Try ID match
	for _, col := range cs.collections {
		if col.ID == nameOrID {
			return col, true
		}
	}
	return nil, false
}

// CreateCollection registers a new resource collection
func (cs *CollectionStore) CreateCollection(col models.Collection) (*models.Collection, error) {
	if err := col.Validate(); err != nil {
		return nil, err
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	key := strings.ToLower(col.Name)
	if _, exists := cs.collections[key]; exists {
		return nil, fmt.Errorf("collection '%s' already exists", col.Name)
	}

	now := time.Now()
	if col.CreatedAt.IsZero() {
		col.CreatedAt = now
	}
	col.UpdatedAt = now

	cs.collections[key] = &col
	_ = cs.saveToFileUnsafe()

	return &col, nil
}

// UpdateCollection updates collection metadata
func (cs *CollectionStore) UpdateCollection(id string, col models.Collection) (*models.Collection, error) {
	if err := col.Validate(); err != nil {
		return nil, err
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	var existingKey string
	var existingCol *models.Collection
	for k, c := range cs.collections {
		if c.ID == id {
			existingKey = k
			existingCol = c
			break
		}
	}

	if existingCol == nil {
		return nil, errors.New("collection not found")
	}

	newKey := strings.ToLower(col.Name)
	if newKey != existingKey {
		if _, exists := cs.collections[newKey]; exists {
			return nil, fmt.Errorf("collection name '%s' is already in use", col.Name)
		}
		delete(cs.collections, existingKey)
	}

	col.ID = id
	col.CreatedAt = existingCol.CreatedAt
	col.UpdatedAt = time.Now()
	if col.Items == nil {
		col.Items = existingCol.Items
	}

	cs.collections[newKey] = &col
	_ = cs.saveToFileUnsafe()

	return &col, nil
}

// DeleteCollection removes a collection
func (cs *CollectionStore) DeleteCollection(nameOrID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(nameOrID))
	if _, exists := cs.collections[key]; exists {
		delete(cs.collections, key)
		_ = cs.saveToFileUnsafe()
		return nil
	}

	for k, col := range cs.collections {
		if col.ID == nameOrID {
			delete(cs.collections, k)
			_ = cs.saveToFileUnsafe()
			return nil
		}
	}

	return errors.New("collection not found")
}

// ListCollections returns all collections
func (cs *CollectionStore) ListCollections() []models.Collection {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	res := make([]models.Collection, 0, len(cs.collections))
	for _, col := range cs.collections {
		res = append(res, *col)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res
}

// ListItems queries items within a collection with filtering, sorting, and pagination
func (cs *CollectionStore) ListItems(collectionName string, queryParams map[string][]string) ([]map[string]interface{}, int, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	key := strings.ToLower(strings.TrimSpace(collectionName))
	col, exists := cs.collections[key]
	if !exists {
		return nil, 0, fmt.Errorf("collection '%s' not found", collectionName)
	}

	// 1. Filter items
	filtered := make([]map[string]interface{}, 0, len(col.Items))
	for _, item := range col.Items {
		if matchItem(item, queryParams) {
			filtered = append(filtered, item)
		}
	}

	totalCount := len(filtered)

	// 2. Sort items
	sortKey := getQueryParam(queryParams, "_sort", "sort")
	sortOrder := strings.ToLower(getQueryParam(queryParams, "_order", "order"))
	if sortOrder == "" {
		sortOrder = "asc"
	}

	if sortKey != "" {
		sort.SliceStable(filtered, func(i, j int) bool {
			valI := filtered[i][sortKey]
			valJ := filtered[j][sortKey]
			cmp := compareValues(valI, valJ)
			if sortOrder == "desc" {
				return cmp > 0
			}
			return cmp < 0
		})
	}

	// 3. Paginate items
	pageStr := getQueryParam(queryParams, "_page", "page")
	limitStr := getQueryParam(queryParams, "_limit", "limit")

	page := 1
	limit := 0

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if limit > 0 {
		start := (page - 1) * limit
		if start >= len(filtered) {
			return []map[string]interface{}{}, totalCount, nil
		}
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		return filtered[start:end], totalCount, nil
	}

	return filtered, totalCount, nil
}

// GetItem retrieves a single item by ID
func (cs *CollectionStore) GetItem(collectionName string, id string) (map[string]interface{}, bool, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	key := strings.ToLower(strings.TrimSpace(collectionName))
	col, exists := cs.collections[key]
	if !exists {
		return nil, false, fmt.Errorf("collection '%s' not found", collectionName)
	}

	for _, item := range col.Items {
		if itemIDStr(item["id"]) == id {
			return item, true, nil
		}
	}

	return nil, false, nil
}

// CreateItem adds a new record to the collection
func (cs *CollectionStore) CreateItem(collectionName string, item map[string]interface{}) (map[string]interface{}, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(collectionName))
	col, exists := cs.collections[key]
	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", collectionName)
	}

	newItem := make(map[string]interface{})
	for k, v := range item {
		newItem[k] = v
	}

	// Ensure ID
	if _, hasID := newItem["id"]; !hasID || newItem["id"] == nil || newItem["id"] == "" {
		newItem["id"] = cs.generateNextID(col)
	}

	now := time.Now().Format(time.RFC3339)
	if _, hasCreated := newItem["createdAt"]; !hasCreated {
		newItem["createdAt"] = now
	}
	newItem["updatedAt"] = now

	col.Items = append(col.Items, newItem)
	col.UpdatedAt = time.Now()
	_ = cs.saveToFileUnsafe()

	return newItem, nil
}

// UpdateItem updates or replaces an existing item
func (cs *CollectionStore) UpdateItem(collectionName string, id string, updateData map[string]interface{}, partial bool) (map[string]interface{}, bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(collectionName))
	col, exists := cs.collections[key]
	if !exists {
		return nil, false, fmt.Errorf("collection '%s' not found", collectionName)
	}

	for i, item := range col.Items {
		if itemIDStr(item["id"]) == id {
			var updatedItem map[string]interface{}

			if partial {
				// PATCH: merge fields
				updatedItem = make(map[string]interface{})
				for k, v := range item {
					updatedItem[k] = v
				}
				for k, v := range updateData {
					if k != "id" && k != "createdAt" {
						updatedItem[k] = v
					}
				}
			} else {
				// PUT: replace fields, preserve ID & createdAt
				updatedItem = make(map[string]interface{})
				for k, v := range updateData {
					updatedItem[k] = v
				}
				updatedItem["id"] = item["id"]
				if createdAt, ok := item["createdAt"]; ok {
					updatedItem["createdAt"] = createdAt
				}
			}

			updatedItem["updatedAt"] = time.Now().Format(time.RFC3339)
			col.Items[i] = updatedItem
			col.UpdatedAt = time.Now()
			_ = cs.saveToFileUnsafe()

			return updatedItem, true, nil
		}
	}

	return nil, false, nil
}

// DeleteItem removes an item by ID
func (cs *CollectionStore) DeleteItem(collectionName string, id string) (bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(collectionName))
	col, exists := cs.collections[key]
	if !exists {
		return false, fmt.Errorf("collection '%s' not found", collectionName)
	}

	for i, item := range col.Items {
		if itemIDStr(item["id"]) == id {
			col.Items = append(col.Items[:i], col.Items[i+1:]...)
			col.UpdatedAt = time.Now()
			_ = cs.saveToFileUnsafe()
			return true, nil
		}
	}

	return false, nil
}

// generateNextID generates sequential integer or random string ID
func (cs *CollectionStore) generateNextID(col *models.Collection) interface{} {
	maxInt := 0
	allInt := true

	for _, item := range col.Items {
		if idVal, exists := item["id"]; exists {
			switch v := idVal.(type) {
			case float64:
				if int(v) > maxInt {
					maxInt = int(v)
				}
			case int:
				if v > maxInt {
					maxInt = v
				}
			case string:
				if num, err := strconv.Atoi(v); err == nil {
					if num > maxInt {
						maxInt = num
					}
				} else {
					allInt = false
				}
			default:
				allInt = false
			}
		}
	}

	if allInt {
		return maxInt + 1
	}

	nBig, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("id_%d", 100000+nBig.Int64())
}

// Helper functions for querying
func matchItem(item map[string]interface{}, queryParams map[string][]string) bool {
	if len(queryParams) == 0 {
		return true
	}

	for paramKey, vals := range queryParams {
		if len(vals) == 0 {
			continue
		}
		val := vals[0]

		// Reserved parameters
		if strings.HasPrefix(paramKey, "_") || paramKey == "page" || paramKey == "limit" || paramKey == "sort" || paramKey == "order" {
			continue
		}

		// Full text search ?q=keyword
		if paramKey == "q" {
			matched := false
			qLower := strings.ToLower(val)
			for _, v := range item {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", v)), qLower) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
			continue
		}

		// Operator checks: _gte, _lte, _gt, _lt, _ne, _like
		if strings.HasSuffix(paramKey, "_gte") {
			field := strings.TrimSuffix(paramKey, "_gte")
			if !compareOp(item[field], val, ">=") {
				return false
			}
			continue
		}
		if strings.HasSuffix(paramKey, "_lte") {
			field := strings.TrimSuffix(paramKey, "_lte")
			if !compareOp(item[field], val, "<=") {
				return false
			}
			continue
		}
		if strings.HasSuffix(paramKey, "_gt") {
			field := strings.TrimSuffix(paramKey, "_gt")
			if !compareOp(item[field], val, ">") {
				return false
			}
			continue
		}
		if strings.HasSuffix(paramKey, "_lt") {
			field := strings.TrimSuffix(paramKey, "_lt")
			if !compareOp(item[field], val, "<") {
				return false
			}
			continue
		}
		if strings.HasSuffix(paramKey, "_ne") {
			field := strings.TrimSuffix(paramKey, "_ne")
			if fmt.Sprintf("%v", item[field]) == val {
				return false
			}
			continue
		}
		if strings.HasSuffix(paramKey, "_like") {
			field := strings.TrimSuffix(paramKey, "_like")
			if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", item[field])), strings.ToLower(val)) {
				return false
			}
			continue
		}

		// Exact match
		if itemVal, ok := item[paramKey]; ok {
			if !strings.EqualFold(fmt.Sprintf("%v", itemVal), val) {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func compareOp(itemVal interface{}, targetStr string, op string) bool {
	fItem, err1 := toFloat(itemVal)
	fTarget, err2 := strconv.ParseFloat(targetStr, 64)

	if err1 == nil && err2 == nil {
		switch op {
		case ">=":
			return fItem >= fTarget
		case "<=":
			return fItem <= fTarget
		case ">":
			return fItem > fTarget
		case "<":
			return fItem < fTarget
		}
	}

	sItem := fmt.Sprintf("%v", itemVal)
	switch op {
	case ">=":
		return sItem >= targetStr
	case "<=":
		return sItem <= targetStr
	case ">":
		return sItem > targetStr
	case "<":
		return sItem < targetStr
	}
	return false
}

func toFloat(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("nil value")
	}
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case string:
		return strconv.ParseFloat(n, 64)
	default:
		return 0, errors.New("unsupported numeric type")
	}
}

func compareValues(a, b interface{}) int {
	fA, errA := toFloat(a)
	fB, errB := toFloat(b)
	if errA == nil && errB == nil {
		if fA < fB {
			return -1
		} else if fA > fB {
			return 1
		}
		return 0
	}
	sA := fmt.Sprintf("%v", a)
	sB := fmt.Sprintf("%v", b)
	return strings.Compare(sA, sB)
}

func itemIDStr(id interface{}) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getQueryParam(params map[string][]string, keys ...string) string {
	for _, k := range keys {
		if vals, ok := params[k]; ok && len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

// File persistence
func (cs *CollectionStore) saveToFileUnsafe() error {
	if cs.persistPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(cs.collections, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cs.persistPath, data, 0644)
}

// LoadFromFile restores collections from JSON file
func (cs *CollectionStore) LoadFromFile(filePath string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var cols map[string]*models.Collection
	if err := json.Unmarshal(bytes, &cols); err != nil {
		return err
	}

	cs.collections = cols
	return nil
}
