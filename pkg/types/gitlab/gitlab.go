package gitlab

type OauthToken struct {
	Token string `json:"access_token"`
}

type GitlabUser struct {
	ID	  	  int           `json:"id"`
	Username  string        `json:"username"`
	Groups    []GitlabGroup `json:"groups"`
	Token 	  string        `json:"-"`
	Host 	  string        `json:"-"`
}

type GitlabGroup struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
}

type GitlabProject struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	PathWithNamespace string          `json:"path_with_namespace"`
	WebURL            string          `json:"web_url"`
	Visibility        string          `json:"visibility"`
	LastActivityAt    string          `json:"last_activity_at"`
	Namespace         GitlabNamespace `json:"namespace"`
}

type GitlabNamespace struct {
	ID   int    `json:"id"`
	Kind string `json:"kind"`
}
