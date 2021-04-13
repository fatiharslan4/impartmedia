CREATE TABLE IF NOT EXISTS questionnaire
(
    questionnaire_id  INT UNSIGNED AUTO_INCREMENT NOT NULL,
    name NVARCHAR(50) NOT NULL,
    version INT UNSIGNED NOT NULL,
    enabled BOOL NOT NULL default false,
    PRIMARY KEY (questionnaire_id),
    UNIQUE(name, version)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS question_type
(
    id NVARCHAR(12) NOT NULL,
    `text` NVARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;


CREATE TABLE IF NOT EXISTS question
(
    question_id INT UNSIGNED AUTO_INCREMENT NOT NULL,
    questionnaire_id  INT UNSIGNED NOT NULL,
    question_name NVARCHAR(255) NOT NULL,
    sort_order INT UNSIGNED NOT NULL,
    `text` NVARCHAR(1024) NOT NULL,
    type_id NVARCHAR(10) NOT NULL,
    PRIMARY KEY (question_id),
    INDEX(questionnaire_id),
    UNIQUE(question_name, questionnaire_id),
    FOREIGN KEY (questionnaire_id) REFERENCES questionnaire (questionnaire_id),
    FOREIGN KEY (type_id) REFERENCES question_type (id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS answer
(
    answer_id INT UNSIGNED AUTO_INCREMENT NOT NULL,
    question_id INT UNSIGNED NOT NULL,
    answer_name NVARCHAR(50) NOT NULL,
    sort_order INT UNSIGNED NOT NULL,
    `text` NVARCHAR(1024) NOT NULL,
    PRIMARY KEY (answer_id),
    UNIQUE(answer_name, question_id),
    FOREIGN KEY (question_id) REFERENCES question (question_id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

INSERT INTO question_type(id, `text`) VALUES ('SINGLE', 'Pick One'), ('MULTIPLE', 'Pick all that apply');
INSERT INTO questionnaire (name, version, enabled) VALUES ('onboarding', 1, true);

SET @onboarding_id = (SELECT LAST_INSERT_ID());


INSERT INTO question (questionnaire_id, question_name, sort_order, text, type_id) VALUES
(@onboarding_id, 'Household', 10, 'What does your household look like?', 'SINGLE'),
(@onboarding_id, 'Dependents', 20, 'Do you have any dependents?', 'MULTIPLE'),
(@onboarding_id, 'Generation', 30, 'Which generation do you belong to?', 'SINGLE'),
(@onboarding_id, 'Gender', 40, 'What is your gender identity?', 'SINGLE'),
(@onboarding_id, 'Race', 50, 'What race/ethnicity best describes you?', 'MULTIPLE'),
(@onboarding_id, 'FinancialGoals', 60, 'What is your primary financial goal?', 'SINGLE');

SET @question_name =  'Household';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Single', 10, 'Single adult'),
(@question_id, 'SingleRoommates', 20, 'Single living with others'),
(@question_id, 'Partner', 30, 'Living with partner'),
(@question_id, 'Married', 40, 'Married'),
(@question_id, 'SharedCustody', 50, 'Shared custody');

SET @question_name =  'Dependents';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'None', 10, 'None'),
(@question_id, 'PreSchool', 20, 'Pre-school children (0-4)'),
(@question_id, 'SchoolAge', 30, 'School age children (5-18)'),
(@question_id, 'PostSchool', 40, 'Post school children (19+)'),
(@question_id, 'Parents', 50, 'Parents'),
(@question_id, 'Other', 60, 'Other family members');

SET @question_name =  'Generation';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'GenZ', 10, 'Gen Z (born after 2001)'),
(@question_id, 'Millennial', 20, 'Millennial (born 1981-2000)'),
(@question_id, 'GenX', 30, 'Gen X (born 1965-1980)'),
(@question_id, 'Boomer', 40, 'Boomer (born 1946-1964)');

SET @question_name =  'Gender';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Woman', 10, 'Woman'),
(@question_id, 'Man', 20, 'Man'),
(@question_id, 'NonBinary', 30, 'Non-binary'),
(@question_id, 'NotListed', 40, 'Not listed');

SET @question_name =  'Race';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'AmIndianAlNative', 10, 'American Indian/Alaskan Native'),
(@question_id, 'AsianPacIslander', 20, 'Asian/Pacific Islander'),
(@question_id, 'Black', 30, 'Black/African American'),
(@question_id, 'Hispanic', 40, 'Hispanic/Latino'),
(@question_id, 'SWAsianNAfrican', 50, 'Southwestern Asian/North African'),
(@question_id, 'White', 60, 'White');

SET @question_name =  'FinancialGoals';
SET @question_id = (select max(question_id) from question
                    where questionnaire_id = @onboarding_id and question_name = @question_name);
INSERT INTO answer (question_id, answer_name, sort_order, text) VALUES
(@question_id, 'Retirement', 10, 'Retirement'),
(@question_id, 'SaveCollege', 20, 'Save for college'),
(@question_id, 'House', 30, 'House down payment'),
(@question_id, 'WealthDivestment', 40, 'Wealth divestment'),
(@question_id, 'GenerationalWealth', 50, 'Generational wealth or legacy');
