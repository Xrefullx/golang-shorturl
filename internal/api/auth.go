package api

import (
	"context"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/service"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type (
	Auth struct {
		crypt AuthCrypt
		svc   service.UserManager
	}
	contextKey string
)

var (
	ContextKeyUserID = contextKey("user-id")
)

func NewAuth(svc service.UserManager) Auth {
	return Auth{
		crypt: *NewAuthCrypt(),
		svc:   svc,
	}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID, err := a.authUser(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID.String())
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a *Auth) authUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	if cookie, errCookie := r.Cookie("token"); errCookie == nil {
		id, err := a.crypt.DecodeToken(cookie.Value)
		if err != nil {
			return uuid.Nil, fmt.Errorf("ошибка установки ключа пользователя:%w", err)
		}

		exist, err := a.svc.Exist(r.Context(), id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("ошибка установки ключа пользователя:%w", err)
		}

		if exist {
			return id, nil
		}
	}

	newUserUUID, newUserToken, err := a.newUser(r.Context())
	if err != nil {
		return uuid.Nil, fmt.Errorf("ошибка установки ключа пользователя:%w", err)
	}

	newCookie := http.Cookie{
		Name:     "token",
		Value:    newUserToken,
		MaxAge:   int((time.Hour * 24 * 30).Seconds()),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &newCookie)

	return newUserUUID, nil
}

func (a *Auth) newUser(ctx context.Context) (uuid.UUID, string, error) {
	newUser, err := a.svc.AddUser(ctx)
	if err == nil {
		token, err := a.crypt.EncodeUUID(newUser.ID)
		if err == nil {

			return newUser.ID, string(token), nil
		}
	}

	return uuid.Nil, "", fmt.Errorf("ошибка гнерации нового пользователя: %w", err)
}
