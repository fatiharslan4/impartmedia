ALTER TABLE hive_admins
    DROP FOREIGN KEY hive_admins_ibfk_1;

ALTER TABLE hive_members
    DROP FOREIGN KEY hive_members_ibfk_1;
    
ALTER TABLE hive_user_demographic
    DROP FOREIGN KEY hive_user_demographic_ibfk_1;
   
ALTER TABLE post
    DROP FOREIGN KEY post_ibfk_1;

delete from hive where hive_id=0;    
ALTER TABLE hive MODIFY COLUMN hive_id BIGINT UNSIGNED AUTO_INCREMENT;

ALTER TABLE hive_admins
     ADD FOREIGN KEY (admin_hive_id) REFERENCES hive (hive_id) ;
     
ALTER TABLE hive_members
     ADD FOREIGN KEY (member_hive_id) REFERENCES hive (hive_id) ;

ALTER TABLE hive_user_demographic
     ADD FOREIGN KEY (hive_id) REFERENCES hive (hive_id) ;

ALTER TABLE post
     ADD FOREIGN KEY (hive_id) REFERENCES hive (hive_id) ;