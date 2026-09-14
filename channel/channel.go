// Package channel contains a minimal background worker example.
package channel

// Worker receives one task, processes it, and sends back the result.
func Worker(tasks chan string, completed chan string) {
	task := <-tasks

	result := "completed: " + task

	completed <- result
}
