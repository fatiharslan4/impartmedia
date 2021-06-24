-- 
-- Dev and ios dev updated without data removal from post_tag with this, 
-- we had issues on those endpoints.
-- updated will take effect in prod and pre-prod endpoints
-- 
-- 
delete from post_tag where tag_id = (select tag_id from tag where name = "Other");
delete from tag where name = "Other";