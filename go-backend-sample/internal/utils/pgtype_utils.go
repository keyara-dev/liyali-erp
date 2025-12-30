package utils

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPgtype converts google/uuid.UUID to pgtype.UUID
func UUIDToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// PgtypeToUUID converts pgtype.UUID to google/uuid.UUID
func PgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

// BoolToPgtype converts bool to pgtype.Bool
func BoolToPgtype(b bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  b,
		Valid: true,
	}
}

// PgtypeToBool converts pgtype.Bool to bool
func PgtypeToBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

// Int32ToPgtype converts int32 to pgtype.Int4
func Int32ToPgtype(i int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: i,
		Valid: true,
	}
}

// PgtypeToInt32 converts pgtype.Int4 to int32
func PgtypeToInt32(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

// StringToPgtype converts string to pgtype.Text
func StringToPgtype(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

// PgtypeToString converts pgtype.Text to string
func PgtypeToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// TimeToPgtype converts time.Time to pgtype.Timestamp
func TimeToPgtype(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

// PgtypeToTime converts pgtype.Timestamp to time.Time
func PgtypeToTime(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// TimestampIsValid checks if pgtype.Timestamp is valid
func TimestampIsValid(t pgtype.Timestamp) bool {
	return t.Valid
}
