alter table user_institutions
    add column deleted_at   DATETIME(3)   NULL default NULL;