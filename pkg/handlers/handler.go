package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	photoslibrary "github.com/denysvitali/go-googlephotos"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	oauthConfig *oauth2.Config
	mux *http.ServeMux
	token *oauth2.Token
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func New(clientId string, clientSecret string, httpPort int) *Handler {
	h := Handler{}
	h.oauthConfig = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{photoslibrary.PhotoslibraryScope},
		RedirectURL: fmt.Sprintf("http://127.0.0.1:%d/auth/google/callback", httpPort),
	}

	mux := http.NewServeMux()
	// Root
	mux.HandleFunc("/",  h.redirectLogin)

	// OauthGoogle
	mux.HandleFunc("/auth/google/login", h.oauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback", h.oauthGoogleCallback)

	h.mux = mux
	return &h
}

func (h *Handler) redirectLogin(w http.ResponseWriter, r *http.Request){
	http.Redirect(w, r, "/auth/google/login", http.StatusTemporaryRedirect)
}

func (h *Handler) GetHandler() *http.ServeMux {
	return h.mux
}

func (h *Handler) oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {

	// Create oauthState cookie
	oauthState := h.generateStateOauthCookie(w)
	u := h.oauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (h *Handler) generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (h *Handler) oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := h.getOauthToken(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	h.token = token
}

func (h *Handler) getOauthToken(code string) (*oauth2.Token, error) {

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	return token, nil
}

func (h *Handler) HasToken() bool {
	return h.token != nil
}

func (h *Handler) GetToken() *oauth2.Token {
	return h.token
}
