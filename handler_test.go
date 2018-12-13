package jparser

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
)

func testHandle(path string, w http.ResponseWriter, req *http.Request, h func(http.ResponseWriter, *http.Request)) {
	r := mux.NewRouter()
	r.HandleFunc(path, h).Methods(req.Method)
	r.ServeHTTP(w, req)
}

func testCheckResponse(loc string, t *testing.T, w *httptest.ResponseRecorder, respStatus int, respBody string) {
	if w.Code != respStatus {
		t.Errorf("[%s]:\twrong StatusCode: got %d, expected %d",
			loc, w.Code, respStatus)
	}

	body, _ := ioutil.ReadAll(w.Result().Body)
	bodyStr := string(body)
	if bodyStr != respBody {
		t.Errorf("[%s]:\twrong Response: got \n%s\n, expected \n%s",
			loc, bodyStr, respBody)
	}
}

func TestHandlerDoSomeWork(t *testing.T) {
	cases := []struct {
		Num      string
		Request  string
		Response string
		Status   int
	}{
		{
			Num:      "1",
			Request:  `{"data":123,"list":[{"item":"a"},{"item":"b"},{"item":"c"}]}`,
			Response: "",
			Status:   http.StatusOK,
		},
		{
			Num:      "2",
			Request:  `test`,
			Response: "",
			Status:   http.StatusOK,
		},
		{
			Num:      "3",
			Request:  "",
			Response: "",
			Status:   http.StatusOK,
		},
	}

	for _, c := range cases {
		h := RequestHandler{
			timer: func() <-chan time.Time {
				t := make(chan time.Time, 1)
				t <- time.Now()
				return t
			},
		}

		url := "/worker"
		req := httptest.NewRequest("GET", url, bytes.NewBuffer([]byte(c.Request)))
		w := httptest.NewRecorder()

		testHandle("/worker", w, req, h.DoSomeWork)
		testCheckResponse("DoSomeWork:"+c.Num, t, w, c.Status, c.Response)
	}
}

func TestHandlerParseJSON(t *testing.T) {
	cases := []struct {
		Num      string
		Request  string
		Response string
		Status   int
	}{
		{
			Num:     "1",
			Request: `{"data":123,"list":[{"item":"a"},{"item":"b"},{"item":"c"}]}`,
			Response: `{
  "data": 123,
  "list": [
    {
      "item": "a"
    },
    {
      "item": "b"
    },
    {
      "item": "c"
    }
  ]
}`,
			Status: http.StatusOK,
		},
		{
			Num:      "2",
			Request:  `[`,
			Response: "json parsing error\n",
			Status:   http.StatusInternalServerError,
		},
	}

	for _, c := range cases {
		h := RequestHandler{
			timer: func() <-chan time.Time {
				t := make(chan time.Time, 1)
				t <- time.Now()
				return t
			},
		}

		url := "/json-parser"
		req := httptest.NewRequest("POST", url, bytes.NewBuffer([]byte(c.Request)))
		req.Header.Add("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testHandle("/json-parser", w, req, h.ParseJSON)
		testCheckResponse("ParseJSON:"+c.Num, t, w, c.Status, c.Response)
	}
}

func TestHandlerParseJSONapi(t *testing.T) {
	cases := []struct {
		Num      string
		Request  string
		Response string
		Status   int
	}{
		{
			Num:     "1",
			Request: `{"data":{"type": "articles","id": "1","attributes": {"title": "test"}}}`,
			Response: `{
  "data": {
    "attributes": {
      "title": "test"
    },
    "id": "1",
    "type": "articles"
  }
}`,
			Status: http.StatusOK,
		},
		{
			Num:      "2",
			Request:  `{"data":123,"list":[{"item":"a"},{"item":"b"},{"item":"c"}]}`,
			Response: "json parsing error\n",
			Status:   http.StatusInternalServerError,
		},
		{
			Num:      "3",
			Request:  `[`,
			Response: "json parsing error\n",
			Status:   http.StatusInternalServerError,
		},
	}

	for _, c := range cases {
		h := RequestHandler{
			timer: func() <-chan time.Time {
				t := make(chan time.Time, 1)
				t <- time.Now()
				return t
			},
		}

		url := "/jsonapi-parser"
		req := httptest.NewRequest("POST", url, bytes.NewBuffer([]byte(c.Request)))
		req.Header.Add("Content-Type", jsonapi.MediaType)
		w := httptest.NewRecorder()

		testHandle("/jsonapi-parser", w, req, h.ParseJSONapi)
		testCheckResponse("ParseJSONapi:"+c.Num, t, w, c.Status, c.Response)
	}
}
