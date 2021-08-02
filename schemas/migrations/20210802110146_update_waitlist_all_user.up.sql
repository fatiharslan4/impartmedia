DELETE hive_members 
FROM hive_members
        Inner  JOIN (
				select member_impart_wealth_id, count(member_impart_wealth_id ) ,max(member_hive_id) as hive_id
				from hive_members
                join user
				ON user.impart_wealth_id = hive_members.member_impart_wealth_id 
				and user.deleted_at is null
				-- and user.blocked=0
				and user.admin=0
				group by member_impart_wealth_id
				having count(member_impart_wealth_id) = 2
        )
    newtable ON hive_members.member_impart_wealth_id = newtable.member_impart_wealth_id
    and hive_members.member_hive_id=newtable.hive_id;
    
    
update hive_members
join user
ON user.impart_wealth_id = hive_members.member_impart_wealth_id 
and user.deleted_at is null
-- and user.blocked=0
and user.admin=0
set member_hive_id =1
where member_hive_id =2

