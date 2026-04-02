package modal

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
)

func makeGroceries(names ...string) []models.GroceryEntry {
	entries := make([]models.GroceryEntry, len(names))
	for i, name := range names {
		entries[i] = models.GroceryEntry{ItemDesc: name}
		entries[i].ID = uint(i + 1)
	}
	return entries
}

func TestBuildGroremoveModalComponents_SingleChunk(t *testing.T) {
	groceries := makeGroceries("Apples", "Bread", "Milk")
	components := buildGroremoveModalComponents(groceries, nil)

	if len(components) != 1 {
		t.Fatalf("expected 1 label, got %d", len(components))
	}
	label := components[0].(discordgo.Label)
	group := label.Component.(discordgo.CheckboxGroup)
	if len(group.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(group.Options))
	}
	if group.Options[0].Value != "1" || group.Options[2].Value != "3" {
		t.Errorf("unexpected option values: %v", group.Options)
	}
}

func TestBuildGroremoveModalComponents_MultipleChunks(t *testing.T) {
	names := make([]string, 25)
	for i := range names {
		names[i] = "Item"
	}
	groceries := makeGroceries(names...)
	components := buildGroremoveModalComponents(groceries, nil)

	// 25 items → 3 groups: [1-10], [11-20], [21-25]
	if len(components) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(components))
	}

	// Check absolute indexes are correct across chunk boundaries
	secondLabel := components[1].(discordgo.Label)
	secondGroup := secondLabel.Component.(discordgo.CheckboxGroup)
	if secondGroup.Options[0].Value != "11" {
		t.Errorf("expected first option of second chunk to be index 11, got %s", secondGroup.Options[0].Value)
	}

	thirdLabel := components[2].(discordgo.Label)
	thirdGroup := thirdLabel.Component.(discordgo.CheckboxGroup)
	if len(thirdGroup.Options) != 5 {
		t.Errorf("expected last chunk to have 5 options, got %d", len(thirdGroup.Options))
	}
	if thirdGroup.Options[4].Value != "25" {
		t.Errorf("expected last option to be index 25, got %s", thirdGroup.Options[4].Value)
	}
}

func TestBuildGroremoveModalComponents_ExactlyOneChunk(t *testing.T) {
	groceries := makeGroceries(make([]string, checkboxGroupMaxOptions)...)
	components := buildGroremoveModalComponents(groceries, nil)

	if len(components) != 1 {
		t.Fatalf("expected exactly 1 label for %d items, got %d", checkboxGroupMaxOptions, len(components))
	}
}

func TestCollectSelectedIndexes_SingleGroup(t *testing.T) {
	components := []discordgo.MessageComponent{
		&discordgo.Label{
			Component: &discordgo.CheckboxGroup{
				CustomID: "groremove_items_0",
				Values:   []string{"2", "5"},
			},
		},
	}
	result := collectSelectedIndexes(components)
	if strings.Join(result, " ") != "2 5" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestCollectSelectedIndexes_MultipleGroups(t *testing.T) {
	components := []discordgo.MessageComponent{
		&discordgo.Label{
			Component: &discordgo.CheckboxGroup{
				CustomID: "groremove_items_0",
				Values:   []string{"3"},
			},
		},
		&discordgo.Label{
			Component: &discordgo.CheckboxGroup{
				CustomID: "groremove_items_1",
				Values:   []string{"11", "14"},
			},
		},
	}
	result := collectSelectedIndexes(components)
	if strings.Join(result, " ") != "3 11 14" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestCollectSelectedIndexes_NoneSelected(t *testing.T) {
	components := buildGroremoveModalComponents(makeGroceries("Apples", "Bread"), nil)
	// Simulate submission with nothing checked (Values stays nil)
	result := collectSelectedIndexes(components)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestBuildGroremoveModalComponents_PreselectedSetsDefault(t *testing.T) {
	groceries := makeGroceries("Apples", "Bread", "Milk")
	components := buildGroremoveModalComponents(groceries, []string{"2"})
	if len(components) != 1 {
		t.Fatalf("expected 1 label, got %d", len(components))
	}
	label := components[0].(discordgo.Label)
	group := label.Component.(discordgo.CheckboxGroup)
	if group.Options[1].Default == nil || !*group.Options[1].Default {
		t.Errorf("expected option 2 to have Default true, got %+v", group.Options[1])
	}
	if (group.Options[0].Default != nil && *group.Options[0].Default) ||
		(group.Options[2].Default != nil && *group.Options[2].Default) {
		t.Errorf("expected only index 2 defaulted, got %+v", group.Options)
	}
}

func TestGroremoveCheckboxOptionLabel_TruncatesLongItemDesc(t *testing.T) {
	long := strings.Repeat("a", 200)
	label := groremoveCheckboxOptionLabel(9, long)
	if utf16Len(label) > discordCheckboxOptionLabelMaxUTF16 {
		t.Fatalf("label UTF-16 length %d > max %d: %q", utf16Len(label), discordCheckboxOptionLabelMaxUTF16, label)
	}
	if !strings.HasPrefix(label, "9. ") {
		t.Fatalf("expected numbered prefix, got %q", label)
	}
}

func TestBuildGroremoveModalComponents_PreselectedAcrossChunks(t *testing.T) {
	names := make([]string, 15)
	for i := range names {
		names[i] = "Item"
	}
	groceries := makeGroceries(names...)
	components := buildGroremoveModalComponents(groceries, []string{"1", "11"})
	if len(components) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(components))
	}
	first := components[0].(discordgo.Label).Component.(discordgo.CheckboxGroup)
	if first.Options[0].Default == nil || !*first.Options[0].Default {
		t.Error("expected first item defaulted in first chunk")
	}
	second := components[1].(discordgo.Label).Component.(discordgo.CheckboxGroup)
	if second.Options[0].Default == nil || !*second.Options[0].Default || second.Options[0].Value != "11" {
		t.Errorf("expected item 11 defaulted in second chunk, got %+v", second.Options[0])
	}
}
