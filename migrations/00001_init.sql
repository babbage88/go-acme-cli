-- +goose Up
-- +goose StatementBegin
CREATE TABLE dns_zones (
    id INTEGER NOT NULL,
    zone_uid TEXT NOT NULL UNIQUE,
    domain_name TEXT NOT NULL,
    PRIMARY KEY (id, zone_uid)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_types (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    record_type TEXT UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO record_types (id, record_type)
VALUES (1, 'A'),
(2, 'AAAA'),
(3, 'MX'),
(4, 'CNAME'),
(6, 'NS'),
(7, 'TXT');
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE dns_records (
    id INTEGER NOT NULL,
    record_uid TEXT NOT NULL UNIQUE,
    zone_id INTEGER NOT NULL,
    zone_uid TEXT NOT NULL UNIQUE,
    type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    content TEXT,
    ttl INTEGER NOT NULL,
    created TEXT DEFAULT (datetime()),
    modified TEXT DEFAULT (datetime()),
    PRIMARY KEY (id, record_uid),
    FOREIGN KEY (type_id) REFERENCES record_type_mapping (record_type_id),
    FOREIGN KEY (zone_uid) REFERENCES dns_zones (zone_uid) ON DELETE CASCADE,
    FOREIGN KEY (zone_id) REFERENCES dns_zones (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_type_mapping (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    record_type_id INTEGER NOT NULL,
    FOREIGN KEY (record_type_id) REFERENCES record_types(id),
    FOREIGN KEY (record_id) REFERENCES dns_records(id) ON DELETE CASCADE 
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_comments (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    comment TEXT,
    FOREIGN KEY (record_id) REFERENCES dns_records(id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_tags (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    tags TEXT,
    FOREIGN KEY (record_id) REFERENCES dns_records(id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE dns_zones;
DROP TABLE dns_records;
DROP TABLE record_types;
DROP TABLE record_comments;
DROP TABLE record_tags;
DROP TABLE record_type_mapping;
-- +goose StatementEnd
