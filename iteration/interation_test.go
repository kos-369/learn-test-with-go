package iteration

import (
	"fmt"
	"testing"
)

// func TestRepeat(t *testing.T) {
// 	repeated := Repeat("a")
// 	expected := "aaaaa"

// 	if repeated != expected {
// 		t.Errorf("expected %q but got %q", expected, repeated)
// 	}
// }

func ExampleRepeat() {
	repeated := Repeat("a")
	fmt.Println(repeated)
	// Output: aaaaa
}

func BenchmarkRepeat(b *testing.B) {
	for b.Loop() {
		Repeat("a")
	}
}
