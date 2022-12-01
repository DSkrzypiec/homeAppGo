package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"homeApp/auth/telegram"
	"homeApp/db"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

const (
	authUserPrefix = "auth/user"
	twoFATimout    = 60 * time.Second
)

var ErrInvalidUsernameOrPass = errors.New("invalid username or password")

// UserAuthenticator should perform user authentication based on their username
// and password comparing it against the database. In case when authentication
// succeeded, then user JWT in string form is returned. In case when
// authentication failed, then error should be non empty.
type UserAuthenticator interface {
	IsUserValid(username, password string) (string, error)
	IsJwtTokenValid(jwtString string) (TokenStatus, error)
	RegenerateJwt(tokenStatus TokenStatus) (string, error)
	Check2FA(username string) (bool, error)
}

type TokenStatus struct {
	IsValid      bool
	UserId       int
	TokenExpUnix int64
}

// UserAuth performs user authentication against user data in the main
// database.
type UserAuth struct {
	DbClient       *db.Client
	JwtSigningKey  []byte
	JwtExpMinutes  int
	TelegramClient *telegram.Client
}

// IsUserValid perform user authentication. Password is combined with user's salt,
// hashed (SHA256) and compared with originally set hashed password.
func (ua UserAuth) IsUserValid(username, password string) (string, error) {
	startTs := time.Now()
	log.Info().Str("username", username).Msgf("[%s] start user authentication", authUserPrefix)

	user, uErr := ua.DbClient.UserByUsername(username)
	if uErr != nil {
		log.Error().Err(uErr).Str("username", username).
			Msgf("[%s] cannot fetch user data", authUserPrefix)
		return "", ErrInvalidUsernameOrPass
	}

	givenPassHashed := ua.prepHashedPassword(password, user.Salt)
	if givenPassHashed != user.PasswordHashed {
		return "", ErrInvalidUsernameOrPass
	}

	userToken, tokenErr := ua.prepJwtString(user.UserId)
	if tokenErr != nil {
		log.Error().Err(tokenErr).Str("username", username).
			Msgf("[%s] cannot sign user token", authUserPrefix)
		return "", tokenErr
	}

	elapsed := time.Since(startTs)
	log.Info().Str("username", username).Dur("duration", elapsed).
		Msgf("[%s] finished user authentication", authUserPrefix)

	return userToken, nil
}

// IsJwtTokenValid verifies whenever given JWT string is valid. That means was
// generated and signed by this backend for existing user.
func (ua UserAuth) IsJwtTokenValid(jwtString string) (TokenStatus, error) {
	startTs := time.Now()
	log.Info().Msgf("[%s] start jwt validation", authUserPrefix)

	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ua.JwtSigningKey, nil
	})
	if err != nil {
		log.Error().Err(err).Str("jwt", jwtString).
			Msgf("[%s] cannot parse given JWT")
		return TokenStatus{}, err
	}

	if !token.Valid {
		invalidErr := fmt.Errorf("token is no longer valid at %v", time.Now())
		log.Error().Err(invalidErr).Msgf("[%s] jwt is no longer valid", authUserPrefix)
		return TokenStatus{}, invalidErr
	}

	claims, claimsOk := token.Claims.(jwt.MapClaims)
	if !claimsOk {
		noClaimsErr := fmt.Errorf("token has not claims")
		log.Error().Err(noClaimsErr).Msgf("[%s] jwt has no claims", authUserPrefix)
		return TokenStatus{}, noClaimsErr
	}

	userId, isOk := claims["userId"]
	if !isOk {
		noUserId := fmt.Errorf("token has not userId claim")
		log.Error().Err(noUserId).Msgf("[%s] jwt has no userId claim", authUserPrefix)
		return TokenStatus{}, noUserId
	}

	tokenExp, teIsOk := claims["exp"]
	if !teIsOk {
		noExp := fmt.Errorf("token has not exp claim for token expiration")
		log.Error().Err(noExp).Msgf("[%s] jwt has no exp claim", authUserPrefix)
		return TokenStatus{}, noExp
	}

	userIdInt := int(userId.(float64))
	tokenExpInt := int64(tokenExp.(float64))
	elapsed := time.Since(startTs)
	log.Info().Int("userId", userIdInt).Int64("tokenExp", tokenExpInt).
		Dur("duration", elapsed).Msgf("[%s] finished jwt validation", authUserPrefix)

	return TokenStatus{
		IsValid:      true,
		UserId:       userIdInt,
		TokenExpUnix: tokenExpInt,
	}, nil
}

// RegenerateJwt takes current session JWT and produce another one with shifted
// timestamp of expiration. Other data stays without any change. This can be
// used to prolong current user session.
func (ua UserAuth) RegenerateJwt(tokenStatus TokenStatus) (string, error) {
	return ua.prepJwtString(tokenStatus.UserId)
}

// Check2FA verifies second step in two-factor authentication which is via
// Telegram channel. This step might be turned off via initial configuration.
// In this case TelegramClient should be nil.
func (ua UserAuth) Check2FA(username string) (bool, error) {
	if ua.TelegramClient == nil { // 2FA is turned off in this case
		return true, nil
	}

	twoFaPassed, twoFaErr := ua.TelegramClient.CheckMessageWithPattern(username, twoFATimout)
	if twoFaErr != nil {
		return false, twoFaErr
	}

	return twoFaPassed, nil
}

func (ua UserAuth) prepJwtString(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().UTC().Add(time.Duration(ua.JwtExpMinutes) * time.Minute).UnixMilli(),
	})

	tokenString, err := token.SignedString(ua.JwtSigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (ua UserAuth) prepHashedPassword(givenPassword, salt string) string {
	givenPassHashed := sha256.Sum256([]byte(givenPassword + salt))
	return fmt.Sprintf("%x", givenPassHashed)
}
