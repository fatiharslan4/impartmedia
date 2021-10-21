package impart

type Message string

var (
	FirstNameRequired      Message = "Firstname required."
	LastNameRequired       Message = "Lastname required."
	NameRequired           Message = "Name required."
	SuperAdminOnly         Message = "Current user does not have the permission."
	HiveRuleCreationFailed Message = "Hive rule creation failed."
	HiveRuleExist          Message = "Hive rule exist."
	HiveRuleNotExist       Message = "No hive rule exist."
	HiveRuleFetchingFailed Message = "Fetching hive rule failed."
	HiveRuletoDbmodel      Message = "Unable to convert hiveRules to  dbmodel."
	HiveRuleNameExist      Message = "Hive rule name already exists."
	HiveRuleLimit          Message = "Hive rule limit must be greater than 0."
)
