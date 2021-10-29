package impart

type Message string

var (
	FirstNameRequired      Message = "First name is required."
	LastNameRequired       Message = "Last name is required."
	NameRequired           Message = "Name required."
	SuperAdminOnly         Message = "Current user does not have the permission."
	HiveRuleCreationFailed Message = "Hive rule creation failed."
	HiveRuleExist          Message = "Could not find the HiveRule."
	HiveRuleNotExist       Message = "No hive rule exist."
	HiveRuleFetchingFailed Message = "Fetching hive rule failed."
	HiveRuletoDbmodel      Message = "Unable to convert hiveRules to  dbmodel."
	HiveRuleNameExist      Message = "Hive rule name already exists."
	HiveRuleLimit          Message = "Hive rule limit must be greater than 0."
	HiveRuleUpdateFailed   Message = "Hive rule updation failed."
	HiveRuleSameStatus     Message = "Hive rule is already in same status."
	HiveRuleEnabled        Message = "The Rule is Enabled."
	HiveRuleDisabled       Message = "The Rule is Disabled."
)
