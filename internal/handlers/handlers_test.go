package handlers

import (
	"bytes"
	"github.com/Xrefullx/golang-shorturl/internal/app"
	"github.com/Xrefullx/golang-shorturl/internal/router"
	"github.com/Xrefullx/golang-shorturl/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_CheckRequestHandler(t *testing.T) {

	testArray := []struct {
		nameTest            string
		method              string
		url                 string
		contentType         string
		body                string
		responseBody        string
		responceContentType string
		responceCode        int

		init func(storage *memory.Maps)
	}{
		{
			nameTest:     "1",
			method:       http.MethodGet,
			url:          "/",
			body:         "",
			responceCode: 400,
		},
		{
			nameTest:     "POST exist URL",
			method:       http.MethodPost,
			url:          "/",
			body:         "https://ya.ru/",
			responceCode: 201,
			init: func(storage *memory.Maps) {
				storage.Save("test", "https://ya.ru/")
			},
		},
	}
	var config app.ServerConfig
	app.EnviromentConfig(&config)
	for _, x := range testArray {
		t.Run((x.nameTest), func(t *testing.T) {
			db := memory.NewStorage()
			if x.init != nil {
				x.init(db)
			}
			sUrl, _ := app.NewShort(db)
			h := &Handler{sUrl: sUrl}
			r := router.CreateRouter(*h)
			request := httptest.NewRequest(x.method, x.url, bytes.NewBuffer([]byte(x.body)))
			if x.contentType != "" {
				request.Header.Set("Content-Type", x.contentType)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			responce := w.Result()
			_, err := io.ReadAll(responce.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer responce.Body.Close()
			require.True(t, x.contentType == "" || responce.Header.Get("Content-Type") == x.contentType, responce.Header.Get("Content-Type"))
			assert.True(t, x.responceCode == 0 || responce.StatusCode == x.responceCode)
		})
	}
}
func TestHandler_SaveHandler(t *testing.T) {
	var config app.ServerConfig
	app.EnviromentConfig(&config)
	long := "https://ya.ru/"
	longURLHeader := "Location"
	db := memory.NewStorage()
	sUrl, _ := app.NewShort(db)
	handler := Handler{sUrl: sUrl, baseUrl: config.Url}
	r := router.CreateRouter(handler)
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(long)))
	request.RemoteAddr = "localhost" + config.Port
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)
	response := w.Result()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	short := string(responseBody)
	request = httptest.NewRequest(http.MethodGet, short, nil)
	request.RemoteAddr = "localhost" + config.Port
	w = httptest.NewRecorder()
	r.ServeHTTP(w, request)
	response = w.Result()
	defer response.Body.Close()
	assert.True(t, response.StatusCode == 307)
	headLocationVal, ok := response.Header[longURLHeader]
	require.True(t, ok)
	assert.Equal(t, long, headLocationVal[0])
}
