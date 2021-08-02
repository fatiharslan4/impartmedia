alter table user
    add column deleted_by_admin BOOL NOT NULL default 0;