update user_demographic set user_count=0;
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

update hive_user_demographic set user_count=0;
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

update hive set created_at='2021-10-06 14:32:55.621' where created_at is null;