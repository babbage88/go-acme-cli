-- +goose Up
-- +goose StatementBegin
CREATE TABLE dns_zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    zone_uid TEXT UNIQUE NOT NULL,
    domain_name TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_type TEXT UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO record_types (id, record_type)
VALUES (1, 'A'),
(2, 'AAAA'),
(3, 'MX'),
(4, 'CNAME'),
(5, 'NS'),
(6, 'TXT');
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE dns_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_uid TEXT UNIQUE NOT NULL,
    zone_id INTEGER,
    zone_uid TEXT NOT NULL,
    type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    content TEXT,
    ttl INTEGER NOT NULL,
    created TEXT DEFAULT (datetime()),
    modified TEXT DEFAULT (datetime()),
    FOREIGN KEY (type_id) REFERENCES record_type_mapping (record_type_id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (zone_uid) REFERENCES dns_zones (zone_uid) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (zone_id) REFERENCES dns_zones (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_type_mapping (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    record_type_id INTEGER NOT NULL,
    FOREIGN KEY (record_type_id) REFERENCES record_types(id),
    FOREIGN KEY (record_id) REFERENCES dns_records(id) ON UPDATE CASCADE ON DELETE CASCADE 
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    comment TEXT,
    FOREIGN KEY (record_id) REFERENCES dns_records(id) ON UPDATE CASCADE ON DELETE CASCADE 
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE record_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
