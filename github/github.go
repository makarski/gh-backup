package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	defaultClient = &http.Client{
		Timeout: time.Second * 5,
	}
)

type (
	Client struct {
		accessToken string
		httpClient  *http.Client
		options     ClientOptions
	}

	ClientOptions struct {
		HTTPClient *http.Client
		Languages  []string
	}

	Repo struct {
		Name         string `json:"name"`
		FullName     string `json:"full_name"`
		SshURL       string `json:"ssh_url"`
		LanguagesURL string `json:"languages_url"`
	}
)

func NewClient(accessToken string, o *ClientOptions) *Client {
	client := &Client{accessToken: accessToken, httpClient: defaultClient}
	if o == nil {
		client.options = ClientOptions{}
		return client
	}

	client.options = *o
	if o.HTTPClient != nil {
		client.httpClient = o.HTTPClient
	}

	return client
}

func (c *Client) ListOrgRepos(orgName string, private bool) ([]Repo, error) {
	reposURL := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   path.Join("orgs", orgName, "repos"),
	}

	if private {
		reposURL.Query().Set("type", "private")
	}

	var repos []Repo
	if err := c.listRepos(reposURL.String(), &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (c *Client) listRepos(urlStr string, repos *[]Repo) error {
	b, nextLink, err := c.getData(urlStr)
	if err != nil {
		return err
	}

	var ghRepos []Repo
	if err := json.Unmarshal(b, &ghRepos); err != nil {
		return err
	}

	if err := c.appendReposFilteredByLanguages(repos, ghRepos); err != nil {
		return err
	}

	if nextLink != "" {
		return c.listRepos(nextLink, repos)
	}

	return nil
}

func (c *Client) appendReposFilteredByLanguages(old *[]Repo, new []Repo) error {
	if len(c.options.Languages) <= 0 {
		*old = append(*old, new...)
		return nil
	}

	for _, repo := range new {
		languages, err := c.getLanguages(repo.LanguagesURL)
		if err != nil {
			return err
		}

		if !languages.containsAny(c.options.Languages...) {
			continue
		}

		*old = append(*old, repo)
	}
	return nil
}

func (c *Client) getLanguages(urlStr string) (Languages, error) {
	b, _, err := c.getData(urlStr)
	if err != nil {
		return nil, err
	}

	languages := make(Languages)
	if err := json.Unmarshal(b, &languages); err != nil {
		return nil, err
	}
	return languages, nil
}

func (c *Client) authGetRequest(urlStr string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.accessToken)
	return req, nil
}

func (c *Client) getData(urlStr string) ([]byte, string, error) {
	req, err := c.authGetRequest(urlStr)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.httpClient.Do(req)
	b, err := handleResponse(resp, err, http.StatusOK)
	if err != nil {
		return nil, "", err
	}

	return b, extractNextLink(resp.Header.Get("Link")), nil
}

func handleResponse(resp *http.Response, err error, okCodes ...int) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	for _, okCode := range okCodes {
		if resp.StatusCode == okCode {
			return b, nil
		}
	}

	if len(b) > 0 {
		return nil, fmt.Errorf("got unexpected error code: %d. %s", resp.StatusCode, b)
	}

	return nil, fmt.Errorf("got unexpected error code: %d", resp.StatusCode)
}

func extractNextLink(linkHeader string) string {
	if len(linkHeader) <= 0 {
		return ""
	}

	linksWithRefs := strings.Split(linkHeader, ", ")
	for _, linkData := range linksWithRefs {
		linkItems := strings.SplitN(linkData, "; ", 2)
		if len(linkItems) < 2 {
			continue
		}

		ref := linkItems[1]
		if len(ref) < 10 {
			continue
		}

		if ref[5:9] == "next" {
			return strings.Trim(linkItems[0], "<>")
		}
	}

	return ""
}
