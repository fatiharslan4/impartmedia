INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 1,question_id,answer_id,0
FROM answer
WHERE answer.answer_id in(66,67,68);

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 2,question_id,answer_id,0
FROM answer
WHERE answer.answer_id in(66,67,68);