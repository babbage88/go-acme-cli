-- name: GetZonesFromDb :many
SELECT id, zone_uid, domain_name FROM dns_zones
ORDER BY id;

-- name: GetZoneIdByDomainName :one
SELECT zone_uid FROM dns_zones WHERE domain_name = ? LIMIT 1;

-- name: GetZoneIdByZoneUid :one
SELECT id FROM dns_zones WHERE zone_uid = ? LIMIT 1;

-- name: GetRecordIdByRecordUid :one
SELECT id FROM dns_records WHERE record_uid = ? LIMIT 1;

-- name: GetRecordsByZoneId :many
SELECT 
    r.id,
    r.record_uid, 
    r.zone_id,
    r.zone_uid,
    r.type_id, 
    r.ttl, 
    r.created, 
    r.modified
FROM dns_records r 
LEFT JOIN record_comments c ON r.id = c.record_id
LEFT JOIN record_tags t ON r.id = t.record_id
WHERE r.zone_uid = ?;

-- name: CreateDnsZone :exec
INSERT INTO dns_zones (zone_uid, domain_name) VALUES(?, ?)
ON CONFLICT (zone_uid) DO UPDATE SET domain_name = excluded.domain_name;

-- name: CreateDnsRecord :one
INSERT OR REPLACE INTO dns_records (record_uid, zone_uid, name, content, type_id, modified, created, ttl)
VALUES(?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (record_uid) DO 
UPDATE SET 
zone_uid = excluded.zone_uid,
name = excluded.name,
content = excluded.content,
type_id = excluded.type_id,
modified = excluded.modified,
created = excluded.created,
ttl = excluded.ttl
RETURNING id, record_uid;


-- name: CreateRecordTypeMapping :exec
INSERT OR REPLACE INTO record_type_mapping (record_id, record_type_id) VALUES(?, ?);

-- name: CreateRecordComment :exec
INSERT OR REPLACE INTO record_comments (record_id, comment) VALUES(?, ?);

-- name: CreateRecordTag :exec
INSERT OR REPLACE INTO record_tags (record_id, tags) VALUES(?, ?);

-- name: UpdateDnsRecordByRecordUid :one
UPDATE dns_records
SET name = ?,
    content = ?,
    type_id = ?,
    ttl = ?,
    modified = datetime()
WHERE record_uid = ?
RETURNING *; 

-- name: UpdateDnsRecordNameByRecordUid :one
UPDATE dns_records
SET name = ?,
    modified = datetime()
WHERE record_uid = ?
RETURNING *; 

-- name: UpdateDnsRecordContentByRecordUid :one
UPDATE dns_records
SET content = ?,
    modified = datetime()
WHERE record_uid = ?
RETURNING *; 

-- name: UpdateDnsRecordTtlByRecordUid :one
UPDATE dns_records
SET ttl = ?,
    modified = datetime()
WHERE record_uid = ?
RETURNING *; 

-- name: UpdateDnsRecordTypeIdByRecordUid :one
UPDATE dns_records
SET type_id = ?,
    modified = datetime()
WHERE record_uid = ?
RETURNING *;

-- name: DeleteRecordByRecordUid :exec
DELETE FROM dns_records WHERE record_uid = ?;

-- name: DeleteRecordByZoneId :exec
DELETE FROM dns_records WHERE zone_uid = ?;
