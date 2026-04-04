//go:build healthcheck

package e2e

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/verzac/grocer-discord-bot/e2e/harness"
)

const (
	GroceryBotIDProd = "815120759680532510" // hard-coding it here since it's not likely to change
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

func TestProdHealthcheck(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	assert := require.New(t)
	listContent := tss.SendAndAwaitReply("!grolist").Content
	assert.Contains(listContent, "1: Chicken")
	assert.Contains(listContent, "2: very delicious milkshake")
	assert.NotContains(listContent, "chicken")

}
