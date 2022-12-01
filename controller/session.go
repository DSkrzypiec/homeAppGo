package controller

import (
	"homeApp/auth"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const contrSessPrefix = "controller/session"

type Session struct {
	UserAuth auth.UserAuthenticator
}

// ProlongHandler handles user session prolongation based on current user JWT.
func (s *Session) ProlongHandler(w http.ResponseWriter, r *http.Request) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start user session prolongation", contrSessPrefix)

	sessCookie, cookieErr := r.Cookie(auth.SessCookieName)
	if cookieErr != nil {
		log.Error().Err(cookieErr).Msgf("[%s] there is no session cookie. Redirect to login", contrSessPrefix)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Should be valid at this point, because the handler is protected by
	// CheckAuth middleware.
	tokenStatus, validErr := s.UserAuth.IsJwtTokenValid(sessCookie.Value)
	if validErr != nil {
		log.Error().Err(validErr).Int("userId", tokenStatus.UserId).Dur("duration", time.Since(startTs)).
			Msgf("[%s] there was error while validating JWT", contrSessPrefix)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !tokenStatus.IsValid {
		log.Error().Int("userId", tokenStatus.UserId).Dur("duration", time.Since(startTs)).
			Msgf("[%s] session expired or credentials are incorrect, redirecting to login", contrSessPrefix)
		http.SetCookie(w, expiredSessionCookie())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	regeneratedJwt, regErr := s.UserAuth.RegenerateJwt(tokenStatus)
	if regErr != nil {
		log.Error().Err(regErr).Int("userId", tokenStatus.UserId).Dur("duration", time.Since(startTs)).
			Msgf("[%s] could not regenerate new JWT", contrSessPrefix)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	regSessCookie := &http.Cookie{
		Name:     auth.SessCookieName,
		Path:     "/",
		Value:    regeneratedJwt,
		Expires:  time.Now().Add(auth.SessCookieExpMinutes * time.Minute),
		HttpOnly: true,
	}
	http.SetCookie(w, regSessCookie)

	log.Info().Dur("duration", time.Since(startTs)).Msgf("[%s] user session prolonged", contrSessPrefix)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// Setting cookie with the same name and expiring time stamp in the past is
// equivalent of deleting the cookie.
func expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     auth.SessCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
}
