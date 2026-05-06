package logger_race

import (
	"sync"
	"testing"

	"github.com/spectra-ai/spectra/logger"
)

// TestNopLogger_ConcurrentCalls verifies that concurrent calls from multiple
// goroutines do not race. This test is meaningful when run with -race flag.
func TestNopLogger_ConcurrentCalls(t *testing.T) {
	l := logger.NewNopLogger()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 4)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			l.Debug("debug msg", "key", "value")
		}()
		go func() {
			defer wg.Done()
			l.Info("info msg", "key", "value")
		}()
		go func() {
			defer wg.Done()
			l.Warn("warn msg", "key", "value")
		}()
		go func() {
			defer wg.Done()
			l.Error("error msg", "key", "value")
		}()
	}

	wg.Wait()
}
