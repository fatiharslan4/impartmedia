SET @onboarding_id =1;
INSERT INTO question (questionnaire_id, question_name, sort_order, text, type_id) VALUES
(@onboarding_id, 'Industry', 90, 'What industry do you work in?', 'SINGLE'),
(@onboarding_id, 'Career', 80, 'What is your career level?', 'SINGLE'),
(@onboarding_id, 'Income', 70, 'What is your income range?', 'SINGLE');


SET @question_name =  'Industry';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Agriculture', 10, 'Agriculture & Forestry/Wildlife'),
(@question_id, 'Business', 20, 'Business & Technology'),
(@question_id, 'Construction', 30, 'Construction/Utilities/Contracting'),
(@question_id, 'Education', 40, 'Education'),
(@question_id, 'EntertainmentGaming', 50, 'Entertainment & Gaming'),
(@question_id, 'Financensurance', 60, 'Finance & Insurance'),
(@question_id, 'FoodHospitality', 70, 'Food & Hospitality'),
(@question_id, 'GovernmentPublicServices', 80, 'Government & Public Services'),
(@question_id, 'HealthServices', 90, 'Health Services & Healthcare'),
(@question_id, 'Legal', 100, 'Legal'),
(@question_id, 'NaturalResources', 101, 'Natural Resources/Environmental'),
(@question_id, 'PersonalProfessionalServices', 102, 'Personal & Professional Services'),
(@question_id, 'RealEstateHousing', 103, 'Real Estate & Housing'),
(@question_id, 'RetaileCommerce', 104, 'Retail & eCommerce'),
(@question_id, 'SafetySecurity', 105, 'Safety & Security'),
(@question_id, 'Transportation', 106, 'Transportation');


SET @question_name =  'Career';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Entrylevel', 10, 'Entry-level'),
(@question_id, 'Midlevel', 20, 'Mid-level'),
(@question_id, 'Management', 30, 'Management'),
(@question_id, 'UpperManagement', 40, 'Upper Management'),
(@question_id, 'BusinessOwner', 50, 'Business Owner'),
(@question_id, 'Other', 60, 'Other');


SET @question_name =  'Income';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Income0', 10, 'Less than $20,000'),
(@question_id, 'Income1', 20, '$20,000 - $44,999'),
(@question_id, 'Income2', 30, '$45,000 - $95,000'),
(@question_id, 'Income3', 40, '$95,000 - 120,000'),
(@question_id, 'Income4', 50, '$120,000+');