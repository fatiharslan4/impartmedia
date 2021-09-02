INSERT INTO user_demographic (answer_id,user_count)
SELECT answer_id,0
FROM answer
WHERE answer.question_id in(7,8,9);

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 1,question_id,answer_id,0
FROM answer
WHERE answer.question_id in(7,8,9);

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 2,question_id,answer_id,0
FROM answer
WHERE answer.question_id in(7,8,9);

