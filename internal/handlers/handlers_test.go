package handlers

import (
	"bytes"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_CheckRequestHandler(t *testing.T) {
	type responce struct {
		ServerResponse int
		body           bool
		responceBody   string
	}
	testArray := []struct {
		name             string
		method           string
		request          string
		body             string
		testShouldReturn responce
	}{
		{
			name:    "1",
			method:  http.MethodGet,
			request: "/testNolinks",
			testShouldReturn: responce{
				ServerResponse: 307,
				body:           false,
			},
		},
		{
			name:    "2",
			method:  http.MethodPost,
			request: "/",
			body:    "",
			testShouldReturn: responce{
				ServerResponse: 500,
				body:           false,
			},
		},
	}
	for _, x := range testArray {
		t.Run((x.name), func(t *testing.T) {
			handlers := Handler{DB: storage.NewStorage()}
			request := httptest.NewRequest(x.method, x.request, bytes.NewBuffer([]byte(x.body)))
			rec := httptest.NewRecorder()
			hend := http.HandlerFunc(handlers.CheckRequestHandler)
			hend.ServeHTTP(rec, request)
			responceServ := rec.Result()
			assert.True(t, responceServ.StatusCode == x.testShouldReturn.ServerResponse, "Жду %d, получаю %d", x.testShouldReturn.ServerResponse, rec.Code)
			defer responceServ.Body.Close()
			responceBody, err := io.ReadAll(responceServ.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.False(t, x.testShouldReturn.body && (x.testShouldReturn.responceBody != string(responceBody)))
		})
	}

}
