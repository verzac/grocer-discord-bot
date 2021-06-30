package e2e

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var tss *testSuiteSession

func TestMain(m *testing.M) {
	tss = setupTestSuite()
	code := m.Run()
	tss.Cleanup()
	os.Exit(code)
}

func setup(tss *testSuiteSession) {
	tss.SendAndAwaitReply("!groclear")
	tss.SendAndAwaitReply("!grobulk\nChicken\nvery delicious milkshake")
}

func TestList(t *testing.T) {
	setup(tss)
	assert := require.New(t)
	listContent := tss.SendAndAwaitReply("!grolist").Content
	assert.Contains(listContent, "1: Chicken")
	assert.Contains(listContent, "2: very delicious milkshake")
	assert.NotContains(listContent, "chicken")
}

func TestRemoveAndClear(t *testing.T) {
	setup(tss)
	assert := require.New(t)
	assert.Contains(tss.SendAndAwaitReply("!groremove 1").Content, "Deleted *Chicken* off your grocery list")
	listContentAfterRemove := tss.SendAndAwaitReply("!grolist").Content
	assert.NotContains(listContentAfterRemove, "Chicken")
	assert.NotContains(listContentAfterRemove, "HEEY WASSUP")
	tss.SendAndAwaitReply("!groclear")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "You have no groceries")
}

func TestEditAndDeets(t *testing.T) {
	setup(tss)
	assert := require.New(t)
	tss.SendAndAwaitReply("!groedit 1 HEEY WASSUP")
	assert.Regexp(
		regexp.MustCompile(fmt.Sprintf("^.*HEEY WASSUP.*(updated by <@%s> \\d+ seconds* ago)", tss.d.State.User.ID)),
		tss.SendAndAwaitReply("!grodeets 1").Content,
	)
}

func TestAdd(t *testing.T) {
	setup(tss)
	assert := require.New(t)
	assert.Contains(tss.SendAndAwaitReply("!gro Chickinz").Content, "Added *Chickinz*")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "3: Chickinz")
}

func TestReset(t *testing.T) {
	setup(tss)
	assert := require.New(t)
	// doing it...
	tss.SendAndAwaitReply("!groreset")
	// confirmed deletion
	tss.AwaitReply()
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "You have no groceries")
}
