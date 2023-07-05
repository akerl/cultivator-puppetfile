package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/akerl/cultivator/plugin"
)

var pattern = regexp.MustCompile(`^([a-z]+) '([\w]+)', '([\d.]+)'$`)

func run(_ string) error {
	return plugin.FindReplace("Puppetfile", pattern, forgeCheck)
}

func forgeCheck(matches []string) string {
	var org string
	switch matches[1] {
	case "hmod":
		org = "halyard"
	case "pmod":
		org = "puppetlabs"
	default:
		return matches[0]
	}

	repo := fmt.Sprintf("%s-%s", org, matches[2])
	version, err := checkModule(repo)
	if err != nil {
		return matches[0]
	}
	return fmt.Sprintf("%s '%s', '%s'", matches[1], matches[2], version)
}

type release struct {
	Version string `json:"version"`
}

type releases struct {
	Results []release `json:"results"`
}

func checkModule(repo string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://forgeapi.puppet.com/v3/releases", nil)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Add("module", repo)
	q.Add("limit", "1")
	q.Add("sort_by", "release_date")
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var rs releases
	err = json.Unmarshal(body, &rs)
	if err != nil {
		return "", err
	}
	if len(rs.Results) != 1 {
		return "", fmt.Errorf("somehow got more or less than one release for %s", repo)
	}
	return rs.Results[0].Version, nil
}

func main() {
	p := plugin.Plugin{
		Commit: plugin.SimpleCommit(
			"Update Puppetfile",
			"cultivator/update-puppetfile",
			"Update modules in Puppetfile based on latest Forge versions",
			"update puppetfile modules",
		),
		Condition: plugin.FileExistsCondition("Puppetfile"),
		Executor:  run,
	}

	p.Run()
}
