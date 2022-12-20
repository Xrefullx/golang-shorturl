package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"io/ioutil"
	"net/http"
)

type Handler struct {
	svc     URLStore
	baseurl string
}

func CreateHandler(svc URLStore, baseURL string) *Handler {
	return &Handler{
		svc:     svc,
		baseurl: baseURL,
	}
}

func (h *Handler) JsonSave(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		h.unsupportedMediaTypeError(w, "Разрешены запросы только в формате JSON!")
		return
	}

	jsBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	defer r.Body.Close()

	incoming := ShortenRequest{}
	if err := json.Unmarshal(jsBody, &incoming); err != nil {
		h.badRequestError(w, "неверный формат JSON")
		return
	}
	if err := incoming.Validate(); err != nil {
		h.badRequestError(w, err.Error())
		return
	}

	shortID, err := h.svc.Save(incoming.SrcURL)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}

	jsResult, err := json.Marshal(ShortenResponse{
		Result: h.baseurl + "/" + shortID,
	})
	if err != nil {
		h.serverError(w, err.Error())
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsResult)
}

func (h *Handler) SaveHandler(w http.ResponseWriter, r *http.Request) {
	search, err := io.ReadAll(r.Body)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	defer r.Body.Close()

	short, err := h.svc.Save(string(search))
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.baseurl + "/" + short))
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	short := chi.URLParam(r, "short")
	long, err := h.svc.Get(short)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	w.Header().Set("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) serverError(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusInternalServerError)
}

func (h *Handler) badRequestError(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusInternalServerError)
}
func (h *Handler) unsupportedMediaTypeError(w http.ResponseWriter, errText string) {
	http.Error(w, errText, http.StatusUnsupportedMediaType)
}

func (h *Handler) notFoundError(w http.ResponseWriter) {
	http.Error(w, "запрашиваемая страница не найдена", http.StatusNotFound)
}
