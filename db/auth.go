package db

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
)

const dbAuthPerfix = "db/auth"

type User struct {
	UserId         int
	Email          string
	Username       string
	PasswordHashed string
	Salt           string
	IsActive       bool
	CreateDate     string
}

// UserByUsername gets user data if given username exists. Otherwise
// non-nil error would be returned.
func (c *Client) UserByUsername(username string) (User, error) {
	log.Info().Str("username", username).Msgf("[%s] reading user info from db", dbAuthPerfix)
	startTs := time.Now()

	var userId, isActive int
	var email, password, salt, createDate string
	row := c.dbConn.QueryRow(userQuery(), username)
	scanErr := row.Scan(&userId, &email, &password, &salt, &isActive, &createDate)

	switch scanErr {
	case sql.ErrNoRows:
		log.Error().Err(scanErr).Str("username", username).
			Msgf("[%s] cannot get user because user does not exist", dbAuthPerfix)
		return User{}, scanErr
	case nil:
		break
	default:
		log.Error().Err(scanErr).Str("username", username).Msgf("[%s] cannot get user info", dbAuthPerfix)
		return User{}, scanErr
	}

	user := User{
		UserId:         userId,
		Email:          email,
		Username:       username,
		PasswordHashed: password,
		Salt:           salt,
		IsActive:       isActive == 1,
		CreateDate:     createDate,
	}
	elapsed := time.Since(startTs)
	log.Info().Str("username", username).Dur("duration", elapsed).Msgf("[%s] finished reading user info", dbAuthPerfix)

	return user, nil
}

func userQuery() string {
	return `
		SELECT
			UserId,
			Email,
			PasswordHashed,
			Salt,
			IsActive,
			CreateDate
		FROM
			users
		WHERE
			Username = ?
		;
	`
}
