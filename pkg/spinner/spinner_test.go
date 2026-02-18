package spinner_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/connect0459/edit-pr-duration/pkg/spinner"
)

func TestSpinner(t *testing.T) {
	t.Run("Start/Stopをデッドロックなく実行できる", func(t *testing.T) {
		var buf bytes.Buffer
		sp := spinner.New("test", &buf)
		sp.Start()
		sp.Stop()
	})

	t.Run("動作中にフレームが出力される", func(t *testing.T) {
		var buf bytes.Buffer
		sp := spinner.New("処理中...", &buf)
		sp.Start()
		time.Sleep(200 * time.Millisecond)
		sp.Stop()

		if buf.Len() == 0 {
			t.Error("スピナーが何も出力していない")
		}
	})

	t.Run("停止後に行がクリアされる", func(t *testing.T) {
		var buf bytes.Buffer
		sp := spinner.New("test", &buf)
		sp.Start()
		time.Sleep(100 * time.Millisecond)
		sp.Stop()

		if !strings.Contains(buf.String(), "\r") {
			t.Error("停止後に行クリアシーケンスが出力されていない")
		}
	})
}
