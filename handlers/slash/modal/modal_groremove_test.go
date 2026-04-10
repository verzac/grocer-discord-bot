package modal

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/utils"
)

func assertGroremoveLabelByteBudget(t *testing.T, label string) {
	t.Helper()
	n := len([]byte(label))
	if n > utils.DiscordCheckboxOptionLabelMaxBytes {
		t.Fatalf("label UTF-8 length %d > max %d: %q", n, utils.DiscordCheckboxOptionLabelMaxBytes, label)
	}
	if !utf8.ValidString(label) {
		t.Fatalf("label is not valid UTF-8: %q", label)
	}
}

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
	assertGroremoveLabelByteBudget(t, label)
	if !strings.HasPrefix(label, "9. ") {
		t.Fatalf("expected numbered prefix, got %q", label)
	}
	if !strings.HasSuffix(label, groremoveTruncationEllipsis) {
		t.Fatalf("expected truncated label to end with ellipsis, got %q", label)
	}
}

func TestGroremoveCheckboxOptionLabel_ExactFitNoEllipsis(t *testing.T) {
	// "1. " is 3 UTF-16 / 3 bytes → 97 characters of ASCII fit exactly under both caps.
	item := strings.Repeat("a", 97)
	label := groremoveCheckboxOptionLabel(1, item)
	assertGroremoveLabelByteBudget(t, label)
	if strings.Contains(label, groremoveTruncationEllipsis) {
		t.Fatalf("should not add ellipsis when item fits exactly: %q", label)
	}
	if label != "1. "+item {
		t.Fatalf("expected full item, got %q", label)
	}
}

func TestGroremoveCheckboxOptionLabel_CJKRespectsUTF8ByteCap(t *testing.T) {
	water := strings.Repeat("\u6c34", 200)
	label := groremoveCheckboxOptionLabel(1, water)
	assertGroremoveLabelByteBudget(t, label)
	if !strings.HasPrefix(label, "1. ") {
		t.Fatalf("expected numbered prefix, got %q", label)
	}
	if !strings.HasSuffix(label, groremoveTruncationEllipsis) {
		t.Fatalf("expected truncated CJK label to end with ellipsis, got %q", label)
	}
}

func TestGroremoveCheckboxOptionLabel_LongEmojiTruncates(t *testing.T) {
	// UTF-8: 4 bytes per emoji; forces aggressive byte truncation vs rune count.
	emoji := "\U0001f600" // 😀
	long := strings.Repeat(emoji, 80)
	label := groremoveCheckboxOptionLabel(2, long)
	assertGroremoveLabelByteBudget(t, label)
	if !strings.HasPrefix(label, "2. ") {
		t.Fatalf("expected numbered prefix, got %q", label)
	}
	if !strings.HasSuffix(label, groremoveTruncationEllipsis) {
		t.Fatalf("expected ellipsis when truncated, got %q", label)
	}
}

func TestGroremoveCheckboxOptionLabel_MixedScriptTruncates(t *testing.T) {
	// Latin + CJK + emoji in one long description.
	part := "café 牛乳 " + "\U0001f6d2" + " "
	long := strings.Repeat(part, 40)
	label := groremoveCheckboxOptionLabel(3, long)
	assertGroremoveLabelByteBudget(t, label)
	if !strings.HasPrefix(label, "3. ") {
		t.Fatalf("expected numbered prefix, got %q", label)
	}
	if !strings.HasSuffix(label, groremoveTruncationEllipsis) {
		t.Fatalf("expected ellipsis when truncated, got %q", label)
	}
}

func TestGroremoveCheckboxOptionLabel_ExactFitCJKPlusASCIINoEllipsis(t *testing.T) {
	// "1. " = 3 bytes; 32× 水 (3 bytes each) + "x" = 97 bytes → total 100, no truncation.
	item := strings.Repeat("\u6c34", 32) + "x"
	label := groremoveCheckboxOptionLabel(1, item)
	assertGroremoveLabelByteBudget(t, label)
	if strings.Contains(label, groremoveTruncationEllipsis) {
		t.Fatalf("should not add ellipsis when item fits exactly: %q", label)
	}
	want := "1. " + item
	if label != want {
		t.Fatalf("want %q, got %q", want, label)
	}
	if len([]byte(label)) != utils.DiscordCheckboxOptionLabelMaxBytes {
		t.Fatalf("expected exactly %d UTF-8 bytes, got %d", utils.DiscordCheckboxOptionLabelMaxBytes, len([]byte(label)))
	}
}

func TestGroremoveCheckboxOptionLabel_ExactFitEmojiNoEllipsis(t *testing.T) {
	// "2. " = 3 bytes; 24 emoji × 4 bytes = 96 → total 99 (room left; add one ASCII for 100).
	emoji := "\U0001f389" // 🎉
	item := strings.Repeat(emoji, 24) + "!"
	label := groremoveCheckboxOptionLabel(2, item)
	assertGroremoveLabelByteBudget(t, label)
	if strings.Contains(label, groremoveTruncationEllipsis) {
		t.Fatalf("should not add ellipsis when item fits exactly: %q", label)
	}
	if label != "2. "+item {
		t.Fatalf("expected full item, got %q", label)
	}
	if len([]byte(label)) != utils.DiscordCheckboxOptionLabelMaxBytes {
		t.Fatalf("expected exactly %d UTF-8 bytes, got %d", utils.DiscordCheckboxOptionLabelMaxBytes, len([]byte(label)))
	}
}

func TestGroremoveCheckboxOptionLabel_LargeIndexPrefixStillValid(t *testing.T) {
	// Wider numbered prefix uses more bytes; remainder should still truncate safely.
	long := strings.Repeat("\u3042", 200) // hiragana あ
	label := groremoveCheckboxOptionLabel(9999, long)
	assertGroremoveLabelByteBudget(t, label)
	if !strings.HasPrefix(label, "9999. ") {
		t.Fatalf("expected prefix, got %q", label)
	}
	if !strings.HasSuffix(label, groremoveTruncationEllipsis) {
		t.Fatalf("expected ellipsis, got %q", label)
	}
}

func TestBuildGroremoveModalComponents_AllOptionLabelsWithinByteBudget(t *testing.T) {
	longJP := strings.Repeat("\u6c34", 50)
	longEmoji := strings.Repeat("\U0001f36a", 30) // 🍪
	names := []string{
		"short",
		longJP,
		"mix " + longEmoji + " tail",
		strings.Repeat("é", 120), // 2 UTF-8 bytes per é
	}
	groceries := makeGroceries(names...)
	components := buildGroremoveModalComponents(groceries, nil)
	for i, comp := range components {
		group := comp.(discordgo.Label).Component.(discordgo.CheckboxGroup)
		for j, opt := range group.Options {
			assertGroremoveLabelByteBudget(t, opt.Label)
			if opt.Label == "" {
				t.Fatalf("component %d option %d: empty label", i, j)
			}
		}
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
