package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"homeApp/db"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

const authUserPrefix = "auth/user"

var ErrInvalidUsernameOrPass = errors.New("invalid username or password")

// UserAuthenticator should perform user authentication based on their username
// and password comparing it against the database. In case when authentication
// succeeded, then user JWT in string form is returned. In case when
// authentication failed, then error should be non empty.
type UserAuthenticator interface {
	IsUserValid(username, password string) (string, error)
	IsJwtTokenValid(jwtString string) (bool, error)
}

// UserAuth performs user authentication against user data in the main
// database.
type UserAuth struct {
	DbClient      *db.Client
	JwtSigningKey []byte
	JwtExpMinutes int
}

// IsUserValid perform user authentication. Password is combined with user's salt,
// hashed (SHA256) and compared with originally set hashed password.
func (ua *UserAuth) IsUserValid(username, password string) (string, error) {
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

	userToken, tokenErr := ua.prepJwtString(user)
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
func (ua *UserAuth) IsJwtTokenValid(jwtString string) (bool, error) {
	startTs := time.Now()
	log.Info().Str("jwt", jwtString).Msgf("[%s] start jwt validation", authUserPrefix)

	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ua.JwtSigningKey, nil
	})
	if err != nil {
		log.Error().Err(err).Str("jwt", jwtString).
			Msgf("[%s] cannot parse given JWT")
		return false, err
	}

	if !token.Valid {
		invalidErr := fmt.Errorf("token is no longer valid at %v", time.Now())
		log.Error().Err(invalidErr).Msgf("[%s] jwt is no longer valid", authUserPrefix)
		return false, invalidErr
	}

	claims, claimsOk := token.Claims.(jwt.MapClaims)
	if !claimsOk {
		noClaimsErr := fmt.Errorf("token has not claims")
		log.Error().Err(noClaimsErr).Msgf("[%s] jwt has no claims", authUserPrefix)
		return false, noClaimsErr
	}

	userId, isOk := claims["userId"]
	if !isOk {
		noUserId := fmt.Errorf("token has not userId claim")
		log.Error().Err(noUserId).Msgf("[%s] jwt has no userId claim", authUserPrefix)
		return false, noUserId
	}

	elapsed := time.Since(startTs)
	log.Info().Str("jwt", jwtString).Int("userId", userId.(int)).
		Dur("duration", elapsed).Msgf("[%s] finished jwt validation", authUserPrefix)

	return true, nil
}

func (ua *UserAuth) prepJwtString(user db.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.UserId,
		"exp":    time.Now().UTC().Add(time.Duration(ua.JwtExpMinutes) * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString(ua.JwtSigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (ua *UserAuth) prepHashedPassword(givenPassword, salt string) string {
	givenPassHashed := sha256.Sum256([]byte(givenPassword + salt))
	return fmt.Sprintf("%x", givenPassHashed)
}
