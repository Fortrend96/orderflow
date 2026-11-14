package collections

import (
	"maps"
	"slices"
	"testing"
)

func TestSetZeroValue(t *testing.T) {
	var set Set[string]
	if set.Len() != 0 || set.Contains("coffee") {
		t.Fatal("zero set should be empty")
	}
	if !set.Add("coffee") || set.Add("coffee") || !set.Add("tea") {
		t.Fatal("Add did not distinguish new values from duplicates")
	}
	if set.Len() != 2 || !set.Contains("tea") {
		t.Fatalf("unexpected set state: len=%d", set.Len())
	}
	set.Clear()
	if set.Len() != 0 || set.Contains("coffee") {
		t.Fatal("Clear did not empty set")
	}
	if !set.Add("coffee") {
		t.Fatal("Set should be usable after Clear")
	}
}

func TestSetSupportsArrayKeys(t *testing.T) {
	var set Set[[8]byte]
	key := [8]byte{'C', 'O', 'F', 'F', 'E', 'E', '0', '1'}
	if !set.Add(key) || !set.Contains(key) || set.Add(key) {
		t.Fatal("unexpected behavior with array keys")
	}
}

func TestUniqueByPreservesFirstOccurrenceAndInput(t *testing.T) {
	type product struct {
		SKU  string
		Name string
	}
	input := []product{
		{SKU: "COFFEE01", Name: "Кофе"},
		{SKU: "TEA00001", Name: "Чай"},
		{SKU: "COFFEE02", Name: "Кофе"},
		{SKU: "MILK0001", Name: "Молоко"},
	}
	got := UniqueBy(input, func(p product) string { return p.Name })
	want := []product{input[0], input[1], input[3]}
	if !slices.Equal(got, want) {
		t.Fatalf("got=%v, want=%v", got, want)
	}
	got[0].Name = "Changed"
	if input[0].Name != "Кофе" {
		t.Fatal("result reused input backing array")
	}
	if len(input) != 4 {
		t.Fatal("input modified")
	}
}

func TestUniqueByPreservesNamedSliceAndNil(t *testing.T) {
	type IDs []int
	var nilIDs IDs
	if UniqueBy(nilIDs, func(n int) int { return n }) != nil {
		t.Fatal("nil input must return nil")
	}
	empty := IDs{}
	if result := UniqueBy(empty, func(n int) int { return n }); result == nil {
		t.Fatal("non-nil empty input must produce non-nil empty output")
	}
	var got IDs = UniqueBy(IDs{1, 2, 1, 3}, func(n int) int { return n })
	if !slices.Equal(got, IDs{1, 2, 3}) {
		t.Fatalf("got %v", got)
	}
}

func TestGroupByPreservesOrderWithinGroups(t *testing.T) {
	type record struct {
		Name  string
		Price int64
	}
	values := []record{
		{Name: "first", Price: 100},
		{Name: "other", Price: 200},
		{Name: "third", Price: 100},
	}
	groups := GroupBy(values, func(p record) int64 { return p.Price })
	if len(groups) != 2 {
		t.Fatalf("groups=%v", groups)
	}
	if !slices.Equal(groups[100], []record{values[0], values[2]}) ||
		!slices.Equal(groups[200], []record{values[1]}) {
		t.Fatalf("order within groups changed: %v", groups)
	}
	groups[100][0].Name = "changed"
	if values[0].Name != "first" {
		t.Fatal("groups changed input")
	}
}

func TestGroupByEmptyInputs(t *testing.T) {
	if groups := GroupBy([]int(nil), func(n int) int { return n }); groups == nil || len(groups) != 0 {
		t.Fatal("nil input should produce initialized empty map")
	}
	if groups := GroupBy([]int{}, func(n int) int { return n }); groups == nil || len(groups) != 0 {
		t.Fatal("empty input should produce initialized empty map")
	}
}

func TestStandardSlicesAndMapsShallowClones(t *testing.T) {
	originalSlice := [][]int{{1, 2}}
	copySlice := slices.Clone(originalSlice)
	copySlice[0][0] = 99
	if originalSlice[0][0] != 99 {
		t.Fatal("slices.Clone is shallow")
	}

	originalMap := map[string][]int{"a": {1, 2}}
	copyMap := maps.Clone(originalMap)
	copyMap["a"][0] = 88
	if originalMap["a"][0] != 88 {
		t.Fatal("maps.Clone is shallow")
	}
	copyMap["b"] = []int{3}
	if _, ok := originalMap["b"]; ok {
		t.Fatal("maps.Clone should use an independent map")
	}
}

func TestSortedMapKeys(t *testing.T) {
	prices := map[int]string{300: "high", 100: "low", 200: "mid"}
	got := slices.Sorted(maps.Keys(prices))
	if !slices.Equal(got, []int{100, 200, 300}) {
		t.Fatalf("got=%v", got)
	}
}

var benchmarkUniqueSink []int

func uniqueLinear(values []int) []int {
	result := make([]int, 0, len(values))
	for _, v := range values {
		if !slices.Contains(result, v) {
			result = append(result, v)
		}
	}
	return result
}

func benchmarkData() []int {
	values := make([]int, 2000)
	for i := range values {
		values[i] = i % 1000
	}
	return values
}

func BenchmarkUniqueBySet(b *testing.B) {
	values := benchmarkData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkUniqueSink = UniqueBy(values, func(n int) int { return n })
	}
}

func BenchmarkUniqueLinear(b *testing.B) {
	values := benchmarkData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkUniqueSink = uniqueLinear(values)
	}
}
