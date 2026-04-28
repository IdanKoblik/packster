package requests

import (
	"fmt"
	"bytes"
	"io"
	"encoding/json"
	"net/http"

	"packster/pkg/types/gitlab"
)

func GitlabOauthToken(client *http.Client, payload map[string]string, gitlabHost string) (*gitlab.OauthToken, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", gitlabHost+"/oauth/token", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res gitlab.OauthToken
	_ = json.Unmarshal(respBody, &res)
	return &res, nil
}

func FetchGitlabUser(client *http.Client, token, gitlabHost string) (*gitlab.GitlabUser, error) {
	req, err := http.NewRequest("GET", gitlabHost+"/api/v4/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res gitlab.GitlabUser
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.ID == 0 {
		return nil, fmt.Errorf("failed to get Gitlab user")
	}

	res.Host = gitlabHost
	res.Token = token

	groups, err := FetchGitlabUserGroups(client, token, gitlabHost)
	if err != nil {
		return nil, err
	}
	res.Groups = groups

	return &res, nil
}

func FetchGitlabProject(client *http.Client, token, gitlabHost string, projectID int) (*gitlab.GitlabProject, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d", gitlabHost, projectID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab project request failed: %s: %s", resp.Status, string(body))
	}

	var project gitlab.GitlabProject
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("unmarshal gitlab project: %w: %s", err, string(body))
	}

	return &project, nil
}

func FetchGitlabGroupProjects(client *http.Client, token, gitlabHost string, groupID, minAccessLevel int) ([]gitlab.GitlabProject, error) {
	url := fmt.Sprintf("%s/api/v4/groups/%d/projects?min_access_level=%d&per_page=100&include_subgroups=true", gitlabHost, groupID, minAccessLevel)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab group projects request failed: %s: %s", resp.Status, string(body))
	}

	var projects []gitlab.GitlabProject
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("unmarshal gitlab projects: %w: %s", err, string(body))
	}

	return projects, nil
}

func FetchGitlabUserGroups(client *http.Client, token, gitlabHost string) ([]gitlab.GitlabGroup, error) {
	req, err := http.NewRequest("GET", gitlabHost+"/api/v4/groups", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab groups request failed: %s: %s", resp.Status, string(body))
	}

	var groups []gitlab.GitlabGroup
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("unmarshal gitlab groups: %w: %s", err, string(body))
	}

	return groups, nil
}
