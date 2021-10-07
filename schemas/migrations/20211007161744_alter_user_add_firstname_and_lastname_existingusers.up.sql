update user

join (select impart_wealth_id, 
       attributes->>'$.name' as name_str,
       SUBSTRING_INDEX(attributes->>'$.name',' ',1) as firstName,
       TRIM( SUBSTR(attributes->>'$.name', LOCATE(' ', attributes->>'$.name')) ) AS lastName,
       attributes->'$.name' as name
       from profile)
profileData 
on user.impart_wealth_id=profileData.impart_wealth_id

SET user.first_name = CASE user.first_name  
WHEN null THEN profileData.firstName
WHEN '' THEN profileData.firstName
ELSE user.first_name
END  ,
user.last_name = CASE user.last_name  
WHEN null THEN profileData.lastName
WHEN '' THEN profileData.lastName
ELSE user.last_name
END  ;
 

ALTER TABLE user MODIFY  first_name NVARCHAR(200) NOT NULL;
ALTER TABLE user MODIFY  last_name NVARCHAR(200) NOT NULL;