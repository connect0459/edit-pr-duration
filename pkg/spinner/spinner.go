package spinner

import (
	"fmt"
	"io"
	"time"
)

const (
	ansiCyan       = "\033[36m"
	ansiReset      = "\033[0m"
	ansiClearLine  = "\r\033[K"
	ansiHideCursor = "\033[?25l"
	ansiShowCursor = "\033[?25h"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner はターミナル上でアニメーションを表示するスピナー
type Spinner struct {
	message string
	output  io.Writer
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// New は新しい Spinner を作成する
func New(message string, output io.Writer) *Spinner {
	return &Spinner{
		message: message,
		output:  output,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

// Start はスピナーのアニメーションを開始する
func (s *Spinner) Start() {
	fmt.Fprint(s.output, ansiHideCursor)
	go func() {
		defer close(s.doneCh)
		for i := 0; ; i++ {
			select {
			case <-s.stopCh:
				fmt.Fprint(s.output, ansiClearLine)
				fmt.Fprint(s.output, ansiShowCursor)
				return
			default:
				frame := frames[i%len(frames)]
				fmt.Fprintf(s.output, "%s%s%s%s %s", ansiClearLine, ansiCyan, frame, ansiReset, s.message)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop はスピナーを停止し、行をクリアする
func (s *Spinner) Stop() {
	close(s.stopCh)
	<-s.doneCh
}
