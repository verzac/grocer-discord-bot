//go:build healthcheck

package e2e

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/verzac/grocer-discord-bot/e2e/harness"
)

const (
	maxHealthcheckRetries     = 2
	healthcheckRestartBackoff = 3 * time.Second
)

var tss *harness.TestSuiteSession

func runHealthcheckAndCaptureAnyPanic(m *testing.M) (code int) {
	defer func() {
		// recover from panic and return non-zero code so that we can retry
		if r := recover(); r != nil {
			log.Printf("prod healthcheck TestMain: panic recovered: %v\n", r)
			tss.Cleanup() // might be a no-op since tss.Cleanup isn't needed for SetupTestSuite, but better to be safe than sorry
			code = 1
		}
	}()
	tss = harness.SetupTestSuite()
	defer tss.Cleanup() // ensure that tss is cleaned up
	code = m.Run()
	return
}

func TestMain(m *testing.M) {
	var code int
	for attempt := 0; attempt <= maxHealthcheckRetries; attempt++ {
		if attempt > 0 {
			log.Printf("prod healthcheck TestMain: retry %d of %d (immediate, fresh Discord session)", attempt, maxHealthcheckRetries)
		}
		code = runHealthcheckAndCaptureAnyPanic(m)
		if code == 0 {
			break
		}
		// sad path: healthcheck failed. retry
		log.Printf("prod healthcheck TestMain: m.Run exited %d (non-zero), attempt %d/%d\n", code, attempt+1, maxHealthcheckRetries+1)
		time.Sleep(healthcheckRestartBackoff)
	}
	// cleanup
	tss.Cleanup()
	os.Exit(code)
}

func setup(tss *harness.TestSuiteSession) {
	defer tss.RecoverFromPanic()
}

func TestProdHealthcheck(t *testing.T) {
	setup(tss)
	defer tss.RecoverFromPanic()
	content := tss.SendAndAwaitReply("!grolist").Content
	possibleResponses := []string{
		"Here's your grocery list:",
		"You have no groceries",
	}
	ok := false
	for _, response := range possibleResponses {
		if strings.Contains(content, response) {
			ok = true
			break
		}
	}
	if !ok {
		t.Errorf("expected a response containing one of %v, got %s", possibleResponses, content)
	}
}
