SET @onboarding_id =1;
INSERT INTO question (questionnaire_id, question_name, sort_order, text, type_id) VALUES
(@onboarding_id, 'EmploymentStatus', 100, 'What best describes your employment status?', 'SINGLE');

SET @last_question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id );

update question set sort_order=70 where question_id=@last_question_id;
update question set sort_order=80 where question_id=7;
update question set sort_order=90 where question_id=8;
update question set sort_order=100 where question_id=9;


SET @question_name =  'EmploymentStatus';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'FullTime', 10, 'Full-time employment'),
(@question_id, 'PartTime', 20, 'Part-time employment'),
(@question_id, 'Unemployed', 30, 'Unemployed'),
(@question_id, 'Self', 40, 'Self-employed'),
(@question_id, 'HomeMaker', 50, 'Home-maker'),
(@question_id, 'Student', 60, 'Student'),
(@question_id, 'Retired', 70, 'Retired');




INSERT INTO user_demographic (answer_id,user_count)
SELECT answer_id,0
FROM answer
WHERE answer.question_id =@last_question_id;

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 1,question_id,answer_id,0
FROM answer
WHERE answer.question_id =@last_question_id;

INSERT INTO hive_user_demographic (hive_id,question_id,answer_id,user_count)
SELECT 2,question_id,answer_id,0
FROM answer
WHERE answer.question_id =@last_question_id;
