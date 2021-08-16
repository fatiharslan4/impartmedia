
CREATE TABLE IF NOT EXISTS hive_user_demographic
(
    hive_id         BIGINT      UNSIGNED                   NOT NULL,
    question_id     INT         UNSIGNED                   NOT NULL,
    answer_id       INT         UNSIGNED                   NOT NULL,
    user_count      INT         UNSIGNED                   NOT NULL  DEFAULT 0,
    PRIMARY KEY (hive_id,question_id,answer_id),
    INDEX (hive_id, question_id,answer_id),
    FOREIGN KEY (hive_id) REFERENCES hive (hive_id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES question (question_id) ,
    FOREIGN KEY (answer_id) REFERENCES answer (answer_id) 
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 1,question_id,answer_id,0
FROM answer;

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 2,question_id,answer_id,0
FROM answer;


UPDATE hive_user_demographic
INNER JOIN
(
   SELECT 
    member_hive_id AS hive, 
    answer.question_id, 
    newtable.answer ,
    count(newtable.answercount) as answercount
    FROM user
    JOIN hive_members 
        ON user.impart_wealth_id=hive_members.member_impart_wealth_id
    JOIN (
                    SELECT user_answers.impart_wealth_id,answer_id AS answer, count(answer_id) AS answercount
                    FROM user_answers
					join  user on user.impart_wealth_id = user_answers.impart_wealth_id
                    where  user.deleted_at IS NULL AND user.blocked=0 
                    GROUP by impart_wealth_id,answer_id
            )
        newtable 
        ON user.impart_wealth_id = newtable.impart_wealth_id
    JOIN answer 
        ON answer.answer_id=newtable.answer
    WHERE user.deleted_at IS NULL AND user.blocked=0 
    GROUP BY hive,newtable.answer
 ) cte_user_demographic
ON  cte_user_demographic.hive = hive_user_demographic.hive_id 
	AND cte_user_demographic.question_id = hive_user_demographic.question_id 
	AND cte_user_demographic.answer = hive_user_demographic.answer_id 
SET 
user_count = cte_user_demographic.answercount;