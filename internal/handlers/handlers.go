package handlers

import (
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"io"
	"net/http"
	"unicode/utf8"
)

type Handler struct {
	DB storage.URLStore
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func (h *Handler) CheckRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		q := trimFirstRune(r.URL.Path)
		long, err := h.DB.Get(q)
		if err != nil {
			h.badRequestError(w)
		}
		w.Header().Set("Location", long)
		w.WriteHeader(http.StatusTemporaryRedirect)

		return
	}

	if r.Method == http.MethodPost {
		long, err := io.ReadAll(r.Body)
		if err != nil {
			h.badRequestError(w)
		}
		short, err := h.DB.Save(string(long))
		if err != nil {
			h.badRequestError(w)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "http://localhost:8080/"+short)
		return
	}
	http.Error(w, "GET and POST", http.StatusBadRequest)
}

func (h *Handler) badRequestError(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusInternalServerError)
}
