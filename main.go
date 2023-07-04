package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

type module struct {
	CurrentRelease release `json:"current_release"`
}

func checkModule(repo string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://forgeapi.puppet.com/v3/modules/%s", repo))
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var m module
	err = json.Unmarshal(body, &m)
	if err != nil {
		return "", err
	}
	return m.CurrentRelease.Version, nil
}

func main() {
	p := plugin.Plugin{
		Commit: plugin.SimpleCommit(
			"Update Puppetfile",
			"update-puppetfile",
			"Update modules in Puppetfile based on latest Forge versions",
			"update puppetfile modules",
		),
		Condition: plugin.FileExistsCondition("Puppetfile"),
		Executor:  run,
	}

	p.Run()
}
