package channel

import "testing"

func TestWorker(t *testing.T) {
	// 仕事を渡すchannelと、結果を受け取るchannelを作る。
	tasks := make(chan string, 2)
	completed := make(chan string)

	// workerを別のgoroutineで動かす。
	go Worker(tasks, completed)

	// workerへ2件の仕事を依頼し、これ以上ないことを伝える。
	tasks <- "send welcome email"
	tasks <- "create thumbnail"
	close(tasks)

	// workerがcompletedを閉じるまで、すべての結果を受け取る。
	var got []string
	for result := range completed {
		got = append(got, result)
	}

	want := []string{
		"completed: send welcome email",
		"completed: create thumbnail",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("result %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
