package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type Handler struct {
	sUrl    URLStore
	baseUrl string
}

func CreateHandler(sUrl URLStore, baseUrl string) *Handler {
	return &Handler{
		sUrl:    sUrl,
		baseUrl: baseUrl,
	}
}

func (h *Handler) JsonSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.badRequestError(w, "Only Post")
		return
	}
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	defer r.Body.Close()

	jsonmap := make(map[string]string)
	if err := json.Unmarshal(jsonBody, &jsonmap); err != nil {
		h.badRequestError(w, "check json")
		return
	}
	searchUrl := jsonmap["url"]
	short, err := h.sUrl.Save(string(searchUrl))
	if err != nil {
		h.badRequestError(w, err.Error())
	}
	jsonResult, err := json.Marshal(struct {
		result string `json:"result"`
	}{result: h.baseUrl + "/" + short})
	if err != nil {
		h.serverError(w, err.Error())
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResult)
}

func (h *Handler) SaveHandler(w http.ResponseWriter, r *http.Request) {
	search, err := io.ReadAll(r.Body)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	defer r.Body.Close()

	short, err := h.sUrl.Save(string(search))
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.baseUrl + "/" + short))
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	short := chi.URLParam(r, "short")
	long, err := h.sUrl.Get(short)
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
