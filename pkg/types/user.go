package types

type User struct {
	ID int
	Username string
	DisplayName string
	SsoID int
	Host string
	Orgs []int
}
