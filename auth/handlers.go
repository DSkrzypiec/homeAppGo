package auth

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	authHandlerPrefix    = "auth/handler"
	SessCookieName       = "session"
	SessCookieExpMinutes = 10
)

type HandlerManager struct {
	UserAuthenticator UserAuthenticator
}

// Login HTTP handler performs user authentication. // TODO more
func (hm *HandlerManager) Login(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	r.ParseForm()
	name := r.FormValue("login")
	pass := r.FormValue("pass")
	log.Info().Str("username", name).Msgf("[%s] start user authentication", authHandlerPrefix)

	userJwt, authErr := hm.UserAuthenticator.IsUserValid(name, pass)
	switch authErr {
	case nil:
		break
	case ErrInvalidUsernameOrPass:
		log.Error().Err(authErr).Str("username", name).Dur("duration", time.Since(startTs)).
			Msgf("[%s] authentication failed - incorrect username or password", authHandlerPrefix)
		http.SetCookie(w, expiredSessionCookie())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	default:
		log.Error().Err(authErr).Str("username", name).Dur("duration", time.Since(startTs)).
			Msgf("[%s] authentication failed - might be a backend error", authHandlerPrefix)
		http.SetCookie(w, expiredSessionCookie())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// TODO: Add 2FA via Telegram optionally

	sessCookie := &http.Cookie{
		Name:     SessCookieName,
		Value:    userJwt,
		Expires:  time.Now().Add(SessCookieExpMinutes * time.Minute),
		HttpOnly: true,
	}
	http.SetCookie(w, sessCookie)
	log.Info().Str("username", name).Dur("duration", time.Since(startTs)).
		Msgf("[%s] finished user authentication - cookie is set", authHandlerPrefix)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// CheckAuth is a middleware for user authentication once the session cookie
// was set. It should be used upon every other API endpoint which shouldn't be
// accessed without being authenticated.
func (hm *HandlerManager) CheckAuth(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		startTs := time.Now()
		log.Info().Msgf("[%s] start user session validation", authHandlerPrefix)

		sessCookie, cookieErr := r.Cookie(SessCookieName)
		if cookieErr != nil {
			log.Error().Err(cookieErr).Msgf("[%s] there is no session cookie. Redirect to login", authHandlerPrefix)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		isValid, userId, validErr := hm.UserAuthenticator.IsJwtTokenValid(sessCookie.Value)
		if validErr != nil {
			log.Error().Err(validErr).Int("userId", userId).Dur("duration", time.Since(startTs)).
				Msgf("[%s] there was error while validating JWT", authHandlerPrefix)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if !isValid {
			log.Error().Int("userId", userId).Dur("duration", time.Since(startTs)).
				Msgf("[%s] session expired or credentials are incorrect, redirecting to login", authHandlerPrefix)
			http.SetCookie(w, expiredSessionCookie())
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		log.Info().Int("userId", userId).Dur("duration", time.Since(startTs)).
			Msgf("[%s] user session validation succeeded", authHandlerPrefix)

		// It's fine! Letting traffic flow
		next(w, r)
	}
}

// Setting cookie with the same name and expiring time stamp in the past is
// equivalent of deleting the cookie.
func expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
}
