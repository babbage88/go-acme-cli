// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package infracli_db

import (
	"database/sql"
)

type DnsRecord struct {
	ID        sql.NullInt64
	RecordUid string
	ZoneID    int64
	ZoneUid   string
	TypeID    int64
	Name      string
	Content   sql.NullString
	Ttl       int64
	Created   sql.NullString
	Modified  sql.NullString
}

type DnsZone struct {
	ID         sql.NullInt64
	ZoneUid    string
	DomainName string
}

type RecordComment struct {
	ID       int64
	RecordID int64
	Comment  sql.NullString
}

type RecordTag struct {
	ID       int64
	RecordID int64
	Tags     sql.NullString
}

type RecordType struct {
	ID         int64
	RecordType string
}

type RecordTypeMapping struct {
	ID           int64
	RecordID     int64
	RecordTypeID int64
}
