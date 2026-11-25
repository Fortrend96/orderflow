package catalog

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

// checkCatalogState compares the indexed Catalog with a simple ordered model.
// The model deliberately avoids reusing Catalog's indexing implementation.
func checkCatalogState(t *testing.T, actions []byte) {
	t.Helper()

	if len(actions) > 128 {
		actions = actions[:128]
	}

	store, err := NewCatalog(0)
	if err != nil {
		t.Fatal(err)
	}

	var ids [4]SKU

	for i := range ids {
		ids[i], err = ParseSKU(
			fmt.Sprintf("ITEM%04d", i),
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	model := make(map[SKU]Product)
	order := make([]SKU, 0, len(ids))

	for step, action := range actions {
		sku := ids[int(action)%len(ids)]

		switch (int(action) / len(ids)) % 5 {
		case 0:
			// Add a new product, or reject a duplicate SKU.
			product, err := NewProduct(
				sku.String(),
				"Товар",
				int64(step+1),
			)
			if err != nil {
				t.Fatal(err)
			}

			err = store.Add(product)

			if _, exists := model[sku]; exists {
				if !errors.Is(err, ErrProductExists) {
					t.Fatalf(
						"step %d: duplicate: %v",
						step,
						err,
					)
				}
			} else {
				if err != nil {
					t.Fatalf(
						"step %d: Add: %v",
						step,
						err,
					)
				}

				model[sku] = product
				order = append(order, sku)
			}

		case 1:
			// Remove, retaining order of remaining products.
			_, exists := model[sku]

			if removed := store.Remove(sku); removed != exists {
				t.Fatalf(
					"step %d: Remove returned %v, want %v",
					step,
					removed,
					exists,
				)
			}

			if exists {
				delete(model, sku)

				order = slices.DeleteFunc(
					order,
					func(current SKU) bool {
						return current == sku
					},
				)
			}

		case 2:
			// Update a price, or reject an unknown SKU.
			price := int64(step + 100)

			err := store.UpdatePrice(sku, price)
			original, exists := model[sku]

			if !exists {
				if !errors.Is(err, ErrProductNotFound) {
					t.Fatalf(
						"step %d: missing update error: %v",
						step,
						err,
					)
				}
			} else {
				if err != nil {
					t.Fatalf(
						"step %d: UpdatePrice: %v",
						step,
						err,
					)
				}

				updated, updateErr := original.WithPrice(price)
				if updateErr != nil {
					t.Fatal(updateErr)
				}

				model[sku] = updated
			}

		case 3:
			// The indexed lookup must agree with the reference model.
			product, found := store.Find(sku)
			want, exists := model[sku]

			if found != exists || (found && product != want) {
				t.Fatalf(
					"step %d: Find %s: product=%v, found=%v",
					step,
					sku,
					product,
					found,
				)
			}

		case 4:
			// A clone has identical values and independent storage.
			clone := store.Clone()

			if !slices.Equal(clone.List(), store.List()) {
				t.Fatalf(
					"step %d: Clone mismatch",
					step,
				)
			}

			if len(order) > 0 {
				first := order[0]

				if err := clone.UpdatePrice(first, 9999); err != nil {
					t.Fatal(err)
				}

				actual, _ := store.Find(first)

				if actual != model[first] {
					t.Fatalf(
						"step %d: changing clone modified original",
						step,
					)
				}
			}
		}

		if store.Count() != len(order) {
			t.Fatalf(
				"step %d: Count=%d, want %d",
				step,
				store.Count(),
				len(order),
			)
		}

		items := store.List()

		if len(items) != len(order) {
			t.Fatalf(
				"step %d: List length mismatch",
				step,
			)
		}

		for i, item := range items {
			if item.SKU() != order[i] ||
				item != model[order[i]] {
				t.Fatalf(
					"step %d: List[%d]=%v, model=%v",
					step,
					i,
					item,
					model[order[i]],
				)
			}
		}

		for _, id := range ids {
			actual, found := store.Find(id)
			want, exists := model[id]

			if found != exists ||
				(found && actual != want) {
				t.Fatalf(
					"step %d: Find %s inconsistent with model",
					step,
					id,
				)
			}
		}

		requireIndexConsistent(t, store)
	}
}

func TestCatalogStateMachine(t *testing.T) {
	cases := []struct {
		name    string
		actions []byte
	}{
		{
			name: "empty",
		},
		{
			name:    "repeated adds",
			actions: []byte{0, 0, 0, 1, 1, 2, 3},
		},
		{
			name: "mixed operations",
			actions: []byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
				10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			},
		},
		{
			name: "reversed sequence",
			actions: []byte{
				19, 18, 17, 16, 15, 14, 13, 12, 11,
				10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkCatalogState(t, tc.actions)
		})
	}
}

func FuzzCatalogStateMachine(f *testing.F) {
	f.Add([]byte{})

	f.Add([]byte{
		0, 0, 4, 8, 12, 16, 5, 9, 13,
	})

	f.Add([]byte{
		3, 2, 1, 0, 7, 6, 5, 4,
		15, 14, 13, 12,
	})

	f.Fuzz(func(t *testing.T, actions []byte) {
		checkCatalogState(t, actions)
	})
}
