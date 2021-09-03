update question_type set text='Pick one.' where id='SINGLE';
update question_type set text='Pick all that apply.' where id='MULTIPLE';
insert question_type (id,text) values('SELECT','Select one.');
Update question set type_id='SELECT' where question_name='Industry';