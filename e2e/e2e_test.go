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
	defer tss.recoverFromPanic()
	code := m.Run()
	tss.Cleanup()
	os.Exit(code)
}

func setup(tss *testSuiteSession) {
	defer tss.recoverFromPanic()
	tss.SendAndAwaitReply("!groclear")
	tss.SendAndAwaitReply("!grobulk\nChicken\nvery delicious milkshake")
}

func TestList(t *testing.T) {
	setup(tss)
	defer tss.recoverFromPanic()
	assert := require.New(t)
	listContent := tss.SendAndAwaitReply("!grolist").Content
	assert.Contains(listContent, "1: Chicken")
	assert.Contains(listContent, "2: very delicious milkshake")
	assert.NotContains(listContent, "chicken")
}

func TestRemoveAndClear(t *testing.T) {
	setup(tss)
	defer tss.recoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!grobulk\nSatay\nNasi padang\nTomato")
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
	defer tss.recoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!groclear")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "You have no groceries")
}

func TestEditAndDeets(t *testing.T) {
	setup(tss)
	defer tss.recoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!groedit 1 HEEY WASSUP")
	assert.Regexp(
		regexp.MustCompile(fmt.Sprintf("^.*HEEY WASSUP.*(updated by <@%s> (\\d+ seconds* ago|just now))", tss.d.State.User.ID)),
		tss.SendAndAwaitReply("!grodeets 1").Content,
	)
}

func TestAdd(t *testing.T) {
	setup(tss)
	defer tss.recoverFromPanic()
	assert := require.New(t)
	assert.Contains(tss.SendAndAwaitReply("!gro Chickinz").Content, "Added *Chickinz*")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "3: Chickinz")
}

func TestReset(t *testing.T) {
	setup(tss)
	defer tss.recoverFromPanic()
	assert := require.New(t)
	tss.SendAndAwaitReply("!groreset")
	assert.Contains(tss.SendAndAwaitReply("!grolist").Content, "You have no groceries")
}
