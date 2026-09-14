// Package channel contains a small background worker example.
package channel

// Worker processes tasks until the tasks channel is closed.
func Worker(tasks chan string, completed chan string) {
	for task := range tasks {
		completed <- "completed: " + task
	}

	close(completed)
}
