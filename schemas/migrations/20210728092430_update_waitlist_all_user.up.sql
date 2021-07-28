update hive_members
join user
ON user.impart_wealth_id = hive_members.member_impart_wealth_id 
and user.deleted_at is null
and user.blocked=0
and user.admin=0
set member_hive_id =1
where member_hive_id =2