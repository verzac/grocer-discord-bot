package handlers

import (
	"strings"
	"testing"

	"github.com/verzac/grocer-discord-bot/models"
)

func groceriesWithIDs(descs ...string) []models.GroceryEntry {
	out := make([]models.GroceryEntry, len(descs))
	for i, d := range descs {
		out[i] = models.GroceryEntry{ItemDesc: d, ID: uint(i + 1)}
	}
	return out
}

func TestPreselectedGroremoveOptionValues_EmptyEntry(t *testing.T) {
	got, err := PreselectedGroremoveOptionValues("", groceriesWithIDs("a"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestPreselectedGroremoveOptionValues_IndexList(t *testing.T) {
	g := groceriesWithIDs("Apples", "Bread", "Milk")
	got, err := PreselectedGroremoveOptionValues("1 3", g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, " ") != "1 3" {
		t.Fatalf("got %v", got)
	}
}

func TestPreselectedGroremoveOptionValues_DedupesDuplicateIndexes(t *testing.T) {
	g := groceriesWithIDs("Apples", "Bread")
	got, err := PreselectedGroremoveOptionValues("1 1", g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "1" {
		t.Fatalf("got %v", got)
	}
}

func TestPreselectedGroremoveOptionValues_NameMatch(t *testing.T) {
	g := groceriesWithIDs("Chicken katsu", "Milk")
	got, err := PreselectedGroremoveOptionValues("katsu", g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "1" {
		t.Fatalf("got %v", got)
	}
}

func TestPreselectedGroremoveOptionValues_OutOfRangeIndex(t *testing.T) {
	g := groceriesWithIDs("a", "b")
	_, err := PreselectedGroremoveOptionValues("99", g, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPreselectedGroremoveOptionValues_UnknownName(t *testing.T) {
	g := groceriesWithIDs("a", "b")
	_, err := PreselectedGroremoveOptionValues("nope", g, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
