package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	store := NewInMemoryPlayerStore()
	server := NewPlayerServer(store)
	player := "Pepepr"

	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetScoreRequest(player))
	assertStatus(t, response.Code, http.StatusOK)

	assertResponseBody(t, response.Body.String(), "3")
}

func TestConcurrentRecordingWins(t *testing.T) {
	store := NewInMemoryPlayerStore()
	server := NewPlayerServer(store)
	player := "Pepper"
	wantedWins := 1000

	var wg sync.WaitGroup
	wg.Add(wantedWins)

	for range wantedWins {
		go func() {
			defer wg.Done()
			server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
		}()
	}
	wg.Wait()

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetScoreRequest(player))
	assertStatus(t, response.Code, http.StatusOK)
	assertResponseBody(t, response.Body.String(), fmt.Sprint(wantedWins))
}
