update answer set text="Less than $25,000" where answer_id=53;
update answer set text="$25,000 - $34,999" where answer_id=54;
update answer set text="$35,000 - $49,999" where answer_id=55;
update answer set text="$50,000 - $74,999" where answer_id=56;
update answer set text="$75,000 - $99,999" where answer_id=57;
update answer set text="$100,000 - $149,999" where answer_id=58;

SET @question_name =  'Income';
SET @onboarding_id =  1;
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Income6', 70, '$150,000 - $199,999'),
(@question_id, 'Income7', 80, '$200,000 - $299,999'),
(@question_id, 'Income8', 90, 'More than $300,000');
