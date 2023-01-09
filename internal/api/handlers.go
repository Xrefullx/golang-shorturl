package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Xrefullx/golang-shorturl/internal/model"
	"github.com/Xrefullx/golang-shorturl/internal/service"
	"github.com/Xrefullx/golang-shorturl/internal/shrterr"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	auth    Auth
	svc     service.URLShortener
	baseURL string
}

func NewHandler(shtSvc service.URLShortener, authSvc service.UserManager, baseURL string) (*Handler, error) {
	return &Handler{
		svc:     shtSvc,
		baseURL: baseURL,
		auth:    NewAuth(authSvc),
	}, nil
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Ping(r.Context()); err != nil {
		h.serverError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	var batch BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	userID := h.getUserIDFromContext(r)
	if userID == uuid.Nil {
		h.serverError(w, "User ID is empty")

		return
	}

	if err := h.svc.DeleteURLList(userID, batch...); err != nil {
		h.serverError(w, err.Error())

		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) SaveBatch(w http.ResponseWriter, r *http.Request) {

	var batch []BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	listToAdd := make(map[string]string, len(batch))
	for _, batchEl := range batch {
		listToAdd[batchEl.ID] = batchEl.URL
	}

	userID := h.getUserIDFromContext(r)

	if err := h.svc.SaveURLList(listToAdd, userID); err != nil {
		h.serverError(w, err.Error())

		return
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "   ")
	if err := encoder.Encode(NewBatchListResponseFromMap(listToAdd, h.baseURL)); err != nil {
		h.serverError(w, err.Error())

		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(buffer.Bytes())
}

func (h *Handler) GetUserUrls(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)

	if userID == uuid.Nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	urlList, err := h.svc.GetUserURLList(r.Context(), userID)
	if err != nil {
		h.serverError(w, err.Error())
		return
	}

	if urlList == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if len(urlList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	jsResult, err := json.Marshal(NewShortenListResponseFromCanonical(urlList, h.baseURL))
	if err != nil {
		h.serverError(w, err.Error())
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsResult))
}

func (h *Handler) SaveURLJSONHandler(w http.ResponseWriter, r *http.Request) {

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

	userID := h.getUserIDFromContext(r)
	shortID, err := h.svc.SaveURL(r.Context(), incoming.SrcURL, userID)

	isConflict := false
	if err != nil {
		shortID, isConflict = processConflictErr(err)
		if !isConflict {
			h.badRequestError(w, err.Error())

			return
		}
	}

	jsResult, err := json.Marshal(ShortenResponse{
		Result: h.baseURL + "/" + shortID,
	})
	if err != nil {
		h.serverError(w, err.Error())
		return
	}

	w.Header().Set("content-type", "application/json")
	if isConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write(jsResult)

}

func processConflictErr(err error) (string, bool) {
	if err != nil && errors.Is(err, &shrterr.ErrorConflictSaveURL{}) {
		conflictErr, _ := err.(*shrterr.ErrorConflictSaveURL)
		return conflictErr.ExistShortURL, true
	}
	return "", false
}

func (h *Handler) SaveURLHandler(w http.ResponseWriter, r *http.Request) {
	// read incoming URL
	srcURL, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.serverError(w, err.Error())
		return
	}
	defer r.Body.Close()

	if string(srcURL) == "" {
		h.badRequestError(w, "нельзя сохранить пустую ссылку")
		return
	}

	userID := h.getUserIDFromContext(r)
	shortID, err := h.svc.SaveURL(r.Context(), string(srcURL), userID)

	isConflict := false
	if err != nil {
		shortID, isConflict = processConflictErr(err)
		if !isConflict {
			h.badRequestError(w, err.Error())

			return
		}
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	if isConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write([]byte(h.baseURL + "/" + shortID))
}

func (h *Handler) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")
	if shortID == "" {
		h.badRequestError(w, "короткая ссылка не может быть пустой")
		return
	}

	storedURL, err := h.svc.GetURL(r.Context(), shortID)
	if err != nil {
		h.badRequestError(w, err.Error())
		return
	}

	if storedURL == (model.ShortURL{}) {
		h.notFoundError(w)
		return
	}

	if storedURL.IsDeleted {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.Header().Set("Location", storedURL.URL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (h *Handler) getUserIDFromContext(r *http.Request) uuid.UUID {
	ctxID := r.Context().Value(ContextKeyUserID)
	if ctxID == nil {
		return uuid.Nil
	}

	userUUID, err := uuid.Parse(ctxID.(string))
	if err == nil {
		return userUUID
	}

	return uuid.Nil
}

func (h *Handler) serverError(w http.ResponseWriter, errText string) {
	http.Error(w, errText, http.StatusInternalServerError)
}

func (h *Handler) badRequestError(w http.ResponseWriter, errText string) {
	http.Error(w, errText, http.StatusBadRequest)
}

func (h *Handler) notFoundError(w http.ResponseWriter) {
	http.Error(w, "запрашиваемая страница не найдена", http.StatusNotFound)
}
