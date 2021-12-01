UPDATE user
set admin=true,
avatar_background='#4D4D4F'
where super_admin=true and deleted_at is null;