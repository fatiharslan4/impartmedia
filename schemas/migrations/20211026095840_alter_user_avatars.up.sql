alter table user
    add column avatar_background  NVARCHAR(100) NOT NULL ;

alter table user
    add column avatar_letter  NVARCHAR(100) NOT NULL ;


UPDATE  user
SET avatar_background=(CASE CEIL(RAND()*3)
              WHEN 1 THEN '#30A096'
              WHEN 2 THEN '#DE750D'
              WHEN 3 THEN '#F4D304'
          END)
where admin=false;

UPDATE  user
SET avatar_background='#4D4D4F'
where admin=true;


UPDATE  user
SET avatar_letter='#FFFFFF';