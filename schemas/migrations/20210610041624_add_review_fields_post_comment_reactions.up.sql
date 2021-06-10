alter table post
    add column reviewed       BOOL            NOT NULL DEFAULT 0,
    add column review_comment NVARCHAR(512)   NULL;

alter table comment
    add column reviewed       BOOL            NOT NULL DEFAULT 0,
    add column review_comment NVARCHAR(512)   NULL;