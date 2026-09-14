package channel

import "testing"

func TestWorker(t *testing.T) {
	// 仕事を渡すchannelと、結果を受け取るchannelを作る。
	tasks := make(chan string)
	completed := make(chan string)

	// workerを別のgoroutineで動かす。
	go Worker(tasks, completed)

	// workerへ仕事を依頼する。
	tasks <- "send welcome email"

	// workerから処理結果を受け取る。
	got := <-completed

	want := "completed: send welcome email"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
