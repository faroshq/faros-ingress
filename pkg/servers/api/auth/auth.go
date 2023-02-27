package auth

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/request"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/faroshq/faros-ingress/pkg/api"
	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/faroshq/faros-ingress/pkg/store"
	utiltls "github.com/faroshq/faros-ingress/pkg/util/tls"
)

// Authenticator authenticator is used to authenticate and handle all authentication related tasks
type Authenticator interface {
	// OIDCLogin will redirect user to OIDC provider
	OIDCLogin(w http.ResponseWriter, r *http.Request)
	// OIDCCallback will handle OIDC callback
	OIDCCallback(w http.ResponseWriter, r *http.Request)
	// Authenticate will authenticate the request if user already exists
	Authenticate(r *http.Request) (authenticated bool, user *models.User, err error)
	// ParseJWTToken will parse the JWT token and return the user
	ParseJWTToken(ctx context.Context, token string) (user *models.User, err error)
}

// Static check
var _ Authenticator = &AuthenticatorImpl{}

type AuthenticatorImpl struct {
	config *config.Config

	oAuthSessions *sessions.CookieStore
	store         store.Store
	provider      *oidc.Provider
	verifier      *oidc.IDTokenVerifier
	redirectURL   string
	client        *http.Client
}

func NewAuthenticator(cfg *config.Config, store store.Store, callbackURLPrefix string) (*AuthenticatorImpl, error) {
	var client *http.Client
	var err error
	ctx := context.Background()

	hostingCoreClient, err := kubernetes.NewForConfig(cfg.ClusterRestConfig)
	if err != nil {
		return nil, err
	}

	secret, err := hostingCoreClient.CoreV1().Secrets(cfg.OIDC.OIDCCASecretNamespace).Get(ctx, cfg.OIDC.OIDCCASecretName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if secret != nil {
		crt, ok := secret.Data["tls.crt"]
		if !ok {
			return nil, errors.New("oidc tls.crt not found in secret")
		}
		key, ok := secret.Data["tls.key"]
		if !ok {
			return nil, errors.New("oidc tls.key not found in secret")
		}
		client, err = httpClientForRootCAs(crt, key)
		if err != nil {
			return nil, err
		}
		ctx = oidc.ClientContext(ctx, client)
	}

	redirectURL := cfg.ExternalAPIURL + callbackURLPrefix

	provider, err := oidc.NewProvider(ctx, cfg.OIDC.OIDCIssuerURL)
	if err != nil {
		return nil, err
	}
	// Create an ID token parser, but only trust ID tokens issued to "example-app"
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.OIDC.OIDCClientID,
	})

	da := &AuthenticatorImpl{
		config:        cfg,
		store:         store,
		verifier:      verifier,
		provider:      provider,
		client:        client,
		redirectURL:   redirectURL,
		oAuthSessions: sessions.NewCookieStore([]byte(cfg.OIDC.OIDCAuthSessionKey)),
	}
	return da, nil
}

func (a *AuthenticatorImpl) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	localRedirect := r.URL.Query().Get("redirect_uri")

	var scopes []string

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Getting the session, it's not an issue if we error here
	session, err := a.oAuthSessions.Get(r, "sess")
	if err != nil {
		// print error
	}

	session.Values["state"] = state
	session.Values["redirect_uri"] = localRedirect
	err = a.oAuthSessions.Save(r, w, session)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed persist state: %q", r.Form), http.StatusBadRequest)
		return
	}

	authCodeURL := ""
	scopes = append(scopes, "openid", "profile", "email")
	if r.FormValue("offline_access") != "yes" {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(state)
	} else {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(state, oauth2.AccessTypeOffline)
	}

	http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
}

func (a *AuthenticatorImpl) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	var (
		token *oauth2.Token
	)

	ctx := oidc.ClientContext(r.Context(), a.client)

	var localRedirect string
	oauth2Config := a.oauth2Config(nil)
	switch r.Method {
	case http.MethodGet:
		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := r.FormValue("error"); errMsg != "" {
			http.Error(w, errMsg+": "+r.FormValue("error_description"), http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, fmt.Sprintf("no code in request: %q", r.Form), http.StatusBadRequest)
			return
		}

		session, err := a.oAuthSessions.Get(r, "sess")
		if err != nil {
			http.Error(w, "no session present", http.StatusBadRequest)
			return
		}

		localRedirect = session.Values["redirect_uri"].(string)

		if state := r.FormValue("state"); state != session.Values["state"] {
			http.Error(w, fmt.Sprintf("expected state %q got %q", session.Values["state"], state), http.StatusBadRequest)
			return
		}
		token, err = oauth2Config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		// Form request from frontend to refresh a token.
		refresh := r.FormValue("refresh_token")
		if refresh == "" {
			http.Error(w, fmt.Sprintf("no refresh_token in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		t := &oauth2.Token{
			RefreshToken: refresh,
			Expiry:       time.Now().Add(-time.Hour),
		}
		var err error
		token, err = oauth2Config.TokenSource(ctx, t).Token()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return
	}

	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: extend
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse claim: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = a.registerOrUpdateUser(ctx, claims.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to register user: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.LoginResponse{
		IDToken:       *idToken,
		RawIDToken:    rawIDToken,
		Email:         claims.Email,
		ServerBaseURL: fmt.Sprintf("%s", a.config.ExternalAPIURL),
	}

	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	base64.StdEncoding.EncodeToString(data)

	localRedirect = localRedirect + "?data=" + base64.StdEncoding.EncodeToString(data)
	http.Redirect(w, r, localRedirect, http.StatusSeeOther)

}

func (a *AuthenticatorImpl) Authenticate(r *http.Request) (authenticated bool, user *models.User, err error) {
	// Trying to authenticate via URL query (websocket for SSH/logs, SSE)
	if urlQueryToken := r.URL.Query().Get("_t"); urlQueryToken != "" {
		user, err = a.ParseJWTToken(r.Context(), urlQueryToken)
		if err != nil {
			return false, nil, err
		}

		// authenticated
		return true, user, nil
	}

	if r.Header.Get("Authorization") == "" {
		return false, nil, nil
	}

	// If it's basic auth (service account), it will have 'Basic' instead of
	// 'Bearer'
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer") {
		return false, nil, nil
	}

	token, err := request.AuthorizationHeaderExtractor.ExtractToken(r)
	if err != nil {
		return false, nil, err
	}

	user, err = a.ParseJWTToken(r.Context(), token)
	if err != nil {
		return false, nil, err
	}

	// authenticated
	return true, user, nil
}

// ParseJWTToken validates token's validity and returns models.User that the token belongs to
func (a *AuthenticatorImpl) ParseJWTToken(ctx context.Context, token string) (user *models.User, err error) {
	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	// TODO: extend
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return a.getUser(ctx, claims.Email)
}

// return an HTTP client which trusts the provided root CAs.
func httpClientForRootCAs(crt, key []byte) (*http.Client, error) {
	c, k, err := utiltls.CertificatePairFromBytes(crt, key)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AddCert(c)

	tlsConfig := &tls.Config{
		RootCAs: pool,
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{
					crt,
				},
				PrivateKey: k,
			},
		},
		ServerName:         "faros",
		InsecureSkipVerify: true,
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}, nil
}

func (a *AuthenticatorImpl) oauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.config.OIDC.OIDCClientID,
		ClientSecret: a.config.OIDC.OIDCClientSecret,
		Endpoint:     a.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  a.redirectURL,
	}
}

// registerOrUpdateUser will register or update user in the system when user is authenticated
func (a *AuthenticatorImpl) registerOrUpdateUser(ctx context.Context, email string) (*models.User, error) {
	current, err := a.getUser(ctx, email)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	}

	if current != nil {
		// no update of any kind for now
		return current, nil
	} else {
		// create the user
		return a.store.CreateUser(ctx, models.User{
			Email: email,
		})
	}
}

func (a *AuthenticatorImpl) getUser(ctx context.Context, email string) (*models.User, error) {
	user, err := a.store.GetUser(ctx, models.User{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}
