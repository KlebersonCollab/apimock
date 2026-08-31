package store_test

import (
	"testing"
	"time"

	"apimock/pkg/models"
	"apimock/pkg/store"
)

func TestCollectionStore(t *testing.T) {
	s := store.NewCollectionStore("")

	t.Run("Create and List Collections", func(t *testing.T) {
		col := models.Collection{
			ID:   "col-products",
			Name: "Products",
		}
		created, err := s.CreateCollection(col)
		if err != nil {
			t.Fatalf("failed to create collection: %v", err)
		}
		if created.Name != "products" {
			t.Errorf("expected lowercase collection name 'products', got %s", created.Name)
		}

		cols := s.ListCollections()
		if len(cols) != 1 || cols[0].Name != "products" {
			t.Errorf("expected 1 collection 'products', got %+v", cols)
		}
	})

	t.Run("Auto CRUD on Collection Items", func(t *testing.T) {
		// Create 3 items
		item1, err := s.CreateItem("products", map[string]interface{}{
			"title":    "Ergonomic Keyboard",
			"price":    150.0,
			"category": "hardware",
			"inStock":  true,
		})
		if err != nil {
			t.Fatalf("failed to create item 1: %v", err)
		}
		if item1["id"] != 1 {
			t.Errorf("expected auto ID 1, got %v", item1["id"])
		}

		item2, _ := s.CreateItem("products", map[string]interface{}{
			"title":    "4K Gaming Monitor",
			"price":    450.0,
			"category": "hardware",
			"inStock":  true,
		})
		if item2["id"] != 2 {
			t.Errorf("expected auto ID 2, got %v", item2["id"])
		}

		_, _ = s.CreateItem("products", map[string]interface{}{
			"title":    "Desk Mat",
			"price":    25.0,
			"category": "accessories",
			"inStock":  false,
		})

		// List all items
		items, total, err := s.ListItems("products", nil)
		if err != nil {
			t.Fatalf("failed to list items: %v", err)
		}
		if total != 3 || len(items) != 3 {
			t.Errorf("expected total 3 items, got %d (len %d)", total, len(items))
		}

		// Filter by exact category
		hwItems, hwTotal, _ := s.ListItems("products", map[string][]string{
			"category": {"hardware"},
		})
		if hwTotal != 2 || len(hwItems) != 2 {
			t.Errorf("expected 2 hardware items, got %d", hwTotal)
		}

		// Filter by price >= 100
		expensiveItems, expTotal, _ := s.ListItems("products", map[string][]string{
			"price_gte": {"100"},
		})
		if expTotal != 2 || len(expensiveItems) != 2 {
			t.Errorf("expected 2 items >= 100, got %d", expTotal)
		}

		// Full-text search ?q=Monitor
		searched, sTotal, _ := s.ListItems("products", map[string][]string{
			"q": {"Monitor"},
		})
		if sTotal != 1 || len(searched) != 1 {
			t.Errorf("expected 1 search result for 'Monitor', got %d", sTotal)
		}

		// Sort by price desc & paginate limit 2
		sorted, sortTotal, _ := s.ListItems("products", map[string][]string{
			"_sort":  {"price"},
			"_order": {"desc"},
			"_page":  {"1"},
			"_limit": {"2"},
		})
		if sortTotal != 3 {
			t.Errorf("expected total count 3, got %d", sortTotal)
		}
		if len(sorted) != 2 {
			t.Fatalf("expected 2 paginated items, got %d", len(sorted))
		}
		if sorted[0]["title"] != "4K Gaming Monitor" || sorted[1]["title"] != "Ergonomic Keyboard" {
			t.Errorf("expected sorted desc order [Monitor (450), Keyboard (150)], got [%v, %v]", sorted[0]["title"], sorted[1]["title"])
		}

		// Get single item by ID
		gotItem, found, err := s.GetItem("products", "1")
		if err != nil || !found {
			t.Fatalf("failed to get item 1: %v", err)
		}
		if gotItem["title"] != "Ergonomic Keyboard" {
			t.Errorf("expected 'Ergonomic Keyboard', got %v", gotItem["title"])
		}

		// Partial update (PATCH) item 1
		patched, found, err := s.UpdateItem("products", "1", map[string]interface{}{
			"price": 139.99,
		}, true)
		if err != nil || !found {
			t.Fatalf("failed to patch item 1: %v", err)
		}
		if patched["price"] != 139.99 || patched["title"] != "Ergonomic Keyboard" {
			t.Errorf("expected title preserved and price updated to 139.99, got %+v", patched)
		}

		// Delete item 1
		deleted, err := s.DeleteItem("products", "1")
		if err != nil || !deleted {
			t.Fatalf("failed to delete item 1: %v", err)
		}

		// Verify deletion
		_, foundAfter, _ := s.GetItem("products", "1")
		if foundAfter {
			t.Error("expected item 1 to be deleted, but was found")
		}
	})
}

func TestCollectionStoreValidation(t *testing.T) {
	s := store.NewCollectionStore("")

	col := models.Collection{
		ID:        "col-users",
		Name:      "users",
		CreatedAt: time.Now(),
	}
	_, err := s.CreateCollection(col)
	if err != nil {
		t.Fatalf("unexpected error creating collection: %v", err)
	}

	// Duplicate create error
	_, errDup := s.CreateCollection(col)
	if errDup == nil {
		t.Error("expected duplicate collection error, got nil")
	}

	// Delete collection
	errDel := s.DeleteCollection("users")
	if errDel != nil {
		t.Fatalf("failed to delete collection: %v", errDel)
	}
	if len(s.ListCollections()) != 0 {
		t.Errorf("expected 0 collections after delete, got %d", len(s.ListCollections()))
	}
}
