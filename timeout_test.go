package jparser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testHandler struct {
}

func (h *testHandler) DoSomeWork(w http.ResponseWriter, r *http.Request)   {}
func (h *testHandler) ParseJSON(w http.ResponseWriter, r *http.Request)    {}
func (h *testHandler) ParseJSONapi(w http.ResponseWriter, r *http.Request) {}

func testRunRequests(reqNumber int, reqLimit float64, check func(<-chan int, <-chan time.Time)) {
	cases := []struct {
		method string
		path   string
	}{
		{
			method: "POST",
			path:   "/json-parser",
		},
		{
			method: "POST",
			path:   "/jsonapi-parser",
		},
	}

	for _, c := range cases {

		s := Server{
			reqLimit: reqLimit,
			handler:  &testHandler{},
		}
		r := s.getRouter()

		status := make(chan int, reqNumber)

		timeout := time.After(1 * time.Second)
		for idx := 0; idx < reqNumber; idx++ {
			go func() {
				request, _ := http.NewRequest(c.method, c.path, strings.NewReader("test"))
				request.RemoteAddr = "127.0.0.1:8000"
				response := httptest.NewRecorder()
				r.ServeHTTP(response, request)
				status <- response.Code
			}()
		}

		for idx := 0; idx < reqNumber; idx++ {
			check(status, timeout)
		}
	}
}

func TestRequestNumberLess(t *testing.T) {

	cases := []struct {
		number int
		limit  float64
	}{
		{
			number: 1,
			limit:  2,
		},
		{
			number: 5,
			limit:  10,
		},
		{
			number: 9,
			limit:  10,
		},
		{
			number: 99,
			limit:  100,
		},
	}

	for _, c := range cases {
		testRunRequests(c.number, c.limit, func(status <-chan int, timeout <-chan time.Time) {
			select {
			case st := <-status:
				if st != http.StatusOK {
					t.Error("failed on HTTP request - should not be rejected")
				}
			case <-timeout:
				t.Fatal("technical failure: test takes more time than expected")
			}
		})
	}
}

func TestRequestNumberEqual(t *testing.T) {

	cases := []struct {
		number int
		limit  float64
	}{
		{
			number: 2,
			limit:  2,
		},
		{
			number: 10,
			limit:  10,
		},
		{
			number: 100,
			limit:  100,
		},
	}

	for _, c := range cases {
		testRunRequests(c.number, c.limit, func(status <-chan int, timeout <-chan time.Time) {
			select {
			case st := <-status:
				if st != http.StatusOK {
					t.Error("failed on HTTP request - should not be rejected")
				}
			case <-timeout:
				t.Fatal("technical failure: test takes more time than expected")
			}
		})
	}
}

func TestRequestNumberMore(t *testing.T) {

	cases := []struct {
		number int
		limit  float64
	}{
		{
			number: 2,
			limit:  1,
		},
		{
			number: 11,
			limit:  10,
		},
		{
			number: 20,
			limit:  10,
		},
		{
			number: 101,
			limit:  100,
		},
	}

	for _, c := range cases {
		rejected := false
		testRunRequests(c.number, c.limit, func(status <-chan int, timeout <-chan time.Time) {
			select {
			case st := <-status:
				if st == http.StatusTooManyRequests {
					rejected = true
				} else if st != http.StatusOK && st != http.StatusTooManyRequests {
					t.Error("failed on HTTP request - wrong response status")
				}
			case <-timeout:
				t.Fatal("technical failure: test takes more time than expected")
			}
		})
		if !rejected {
			t.Error("failed on HTTP request - no requests was rejected")
		}
	}
}
