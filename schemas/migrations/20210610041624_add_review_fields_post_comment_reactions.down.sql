alter table post
    drop column reviewed,
    drop column review_comment;


alter table comment
    drop column reviewed,
    drop column review_comment;
