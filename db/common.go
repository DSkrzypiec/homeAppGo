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

func toNullInt(n *int) sql.NullInt32 {
	if n == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Valid: true, Int32: int32(*n)}
}
