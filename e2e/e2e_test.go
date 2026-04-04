//go:build integration

package e2e

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/verzac/grocer-discord-bot/e2e/harness"
)

var tss *harness.TestSuiteSession

func TestMain(m *testing.M) {
	tss = harness.SetupTestSuite()
	defer tss.RecoverFromPanic()
	code := m.Run()
	tss.Cleanup()
	os.Exit(code)
}

func setup(tss *harness.TestSuiteSession) {
	defer tss.RecoverFromPanic()
	tss.SendAndAwaitReply("!groclear")
	tss.SendAndAwaitReply("!grobulk\nChicken\nvery delicious milkshake")
}

func TestList(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	listContent := tss.SendAndAwaitReply("!grolist").Content
	assert.Contains(listContent, "1: Chicken")
	assert.Contains(listContent, "2: very delicious milkshake")
	assert.NotContains(listContent, "chicken")
}

func TestRemoveAndClear(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!grobulk Satay\nNasi padang\nTomato")
	// note that we can just do a grobulk here because we assume that the bot is going to set the grobulk append flag to true for existing guilds
	// do feel free to replace the above with the below if one day we change this init behaviour
	// tss.SendAndAwaitReply("!gro Satay")
	// tss.SendAndAwaitReply("!gro Nasi padang")
	// tss.SendAndAwaitReply("!gro Tomato")
	// test multiple deletes
	assert.Contains(tss.SendAndAwaitReply("!groremove 1 2").Content, "Deleted *Chicken*, and *very delicious milkshake* off your grocery list")
	listContentAfterRemove := tss.SendAndAwaitReply("!grolist").Content
	assert.NotContains(listContentAfterRemove, "Chicken")
	assert.NotContains(listContentAfterRemove, "very delicious milkshake")
	// test name deletes
	// test partial end
	tss.SendAndAwaitReply("!groremove ay")
	assert.NotContains(tss.SendAndAwaitReply("!grolist").Content, "Satay")
	// test partial middle and weird casing
	tss.SendAndAwaitReply("!groremove si pADA")
	assert.NotContains(tss.SendAndAwaitReply("!grolist").Content, "Nasi padang")
	// test partial start
	tss.SendAndAwaitReply("!groremove toMA")
	assert.NotContains(tss.SendAndAwaitReply("!grolist").Content, "Tomato")
}

func TestClear(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!groclear")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "You have no groceries")
}

func TestEditAndDeets(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!groedit 1 HEEY WASSUP")
	assert.Regexp(
		regexp.MustCompile(fmt.Sprintf("^.*HEEY WASSUP.*(updated by <@%s> (\\d+ seconds* ago|just now))", tss.ClientUserID())),
		tss.SendAndAwaitReply("!grodeets 1").Content,
	)
}

func TestAdd(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	assert.Contains(tss.SendAndAwaitReply("!gro Chickinz").Content, "Added *Chickinz*")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "3: Chickinz")
}

func TestReset(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!groreset")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "You have no groceries")
}
