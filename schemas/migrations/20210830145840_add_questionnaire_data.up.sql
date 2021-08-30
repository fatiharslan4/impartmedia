SET @onboarding_id =1
INSERT INTO question (questionnaire_id, question_name, sort_order, text, type_id) VALUES
(@onboarding_id, 'Industry', 90, 'What industry do you work in?', 'SINGLE'),
(@onboarding_id, 'Career', 80, 'What is your career level?', 'MULTIPLE')
(@onboarding_id, 'Income', 70, 'What is your income range?', 'SINGLE'),;


SET @question_name =  'Career';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Entrylevel', 10, 'Entry-level'),
(@question_id, 'Midlevel', 20, 'Mid-level'),
(@question_id, 'Management', 30, 'Management'),
(@question_id, 'UpperManagement', 40, 'Upper Management'),
(@question_id, 'BusinessOwner', 50, 'Business Owner')
(@question_id, 'Other', 60, 'Other');


SET @question_name =  'Income';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Single', 10, 'Less than $20,000'),
(@question_id, 'SingleRoommates', 20, '$20,000 - $44,999'),
(@question_id, 'Partner', 30, '$45,000 - $95,000'),
(@question_id, 'Married', 40, '$95,000 - 120,000'),
(@question_id, 'SharedCustody', 50, '$120,000+');