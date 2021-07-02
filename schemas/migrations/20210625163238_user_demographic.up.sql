DROP TABLE IF EXISTS user_demographic;

CREATE TABLE IF NOT EXISTS user_demographic(
    demographic_id  BIGINT   UNSIGNED AUTO_INCREMENT    NOT NULL,
    answer_id       INT      UNSIGNED                   NOT NULL,
    user_count      INT      NOT NULL                   DEFAULT 0,
    PRIMARY KEY (demographic_id),
    FOREIGN KEY (answer_id) REFERENCES answer (answer_id) 
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

INSERT INTO user_demographic (answer_id,user_count)
SELECT answer_id,0
FROM answer;


UPDATE user_demographic
INNER JOIN
(
SELECT answer_id,count(answer_id) as usercount
 FROM user_answers
 join user on user.impart_wealth_id=user_answers.impart_wealth_id 
 and user.deleted_at is null
 and user.blocked=0
 group by answer_id
 ) cte_user_demographic
ON cte_user_demographic.answer_id = user_demographic.answer_id 
SET 
user_count = cte_user_demographic.usercount;