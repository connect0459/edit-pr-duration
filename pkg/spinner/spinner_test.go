package spinner_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/connect0459/edit-pr-duration/pkg/spinner"
)

func TestSpinner(t *testing.T) {
	t.Run("Start and Stop run without deadlock", func(t *testing.T) {
		var buf bytes.Buffer
		sp := spinner.New("test", &buf)
		sp.Start()
		sp.Stop()
	})

	t.Run("frames are written while running", func(t *testing.T) {
		var buf bytes.Buffer
		sp := spinner.New("Processing...", &buf)
		sp.Start()
		time.Sleep(200 * time.Millisecond)
		sp.Stop()

		if buf.Len() == 0 {
			t.Error("spinner wrote no output")
		}
	})

	t.Run("line is cleared after stopping", func(t *testing.T) {
		var buf bytes.Buffer
		sp := spinner.New("test", &buf)
		sp.Start()
		time.Sleep(100 * time.Millisecond)
		sp.Stop()

		if !strings.Contains(buf.String(), "\r") {
			t.Error("line-clear sequence was not written after stopping")
		}
	})
}
