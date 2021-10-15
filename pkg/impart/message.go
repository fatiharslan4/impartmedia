package impart

type Message string

var (
	FirstNameRequired Message = "Firstname required."
	LastNameRequired  Message = "Lastname required."
	NameRequired      Message = "Name required."
)
