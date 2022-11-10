package db

import "database/sql"

func toNullString(x *string) sql.NullString {
	if x == nil {
		return sql.NullString{Valid: false}
	}
	if *x == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{Valid: true, String: *x}
}
