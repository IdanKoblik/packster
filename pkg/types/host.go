package types

type HostType int

const (
	Gitlab HostType = iota
	Github
)

var gitName = map[HostType]string{
    Github: "github",
    Gitlab: "gitlab",
}

func (ss HostType) String() string {
    return gitName[ss]
}

type Host struct {
	Id 		      int
	Url 		  string
	Type 		  HostType
	ApplicationId string
	Secret		  string
	Orgs 		  []int
}
