package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/concourse/concourse/fly/rc"
)

func main() {
	err := mainE(os.Args[1:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

func mainE(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accessing arguments: expect exactly one argument, but got %d", len(args))
	}

	u, err := url.Parse(args[0])

	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	urlMap := parseUrlPath(u.Path)

	target, teamName, err := rc.LoadTargetFromURL(fmt.Sprintf("%s://%s", u.Scheme, u.Host), urlMap["teams"], false)

	if err != nil {
		return err
	}

	err = target.Validate()

	if err != nil {
		return err
	}

	pipeline := urlMap["pipelines"]

	pipelineWithInstanceVars, err := pipelineWithInstanceVars(pipeline, u.Query())

	if err != nil {
		return err
	}

	instanceVars, err := instanceVars(u.Query())

	if err != nil {
		return err
	}

	fmt.Printf("fly --target %s get-pipeline --pipeline %s --team %s", teamName, pipelineWithInstanceVars, target.Team().Name())
	fmt.Print(" | vipe | ")
	fmt.Printf("fly --target %s set-pipeline --pipeline %s --team %s --config - %s", teamName, pipeline, target.Team().Name(), instanceVars)
	fmt.Println()

	// TODO launch it

	return nil
}

func pipelineWithInstanceVars(pipeline string, query url.Values) (string, error) {
	var pipelineWithInstanceVars strings.Builder

	pipelineWithInstanceVars.WriteString(pipeline)

	if len(query) > 0 {
		pipelineWithInstanceVars.WriteString("/")
	}

	for k, v := range query {
		if !strings.HasPrefix(k, "vars.") {
			continue
		}

		if len(v) > 1 {
			return "", fmt.Errorf("parsing instance variables: expecting ecactly one value for %s, but found %d", k, len(v))
		}

		pipelineWithInstanceVars.WriteString(strings.TrimPrefix(k, "vars."))
		pipelineWithInstanceVars.WriteString(":")
		pipelineWithInstanceVars.WriteString(v[0])

		// TODO Add comma if more than one
	}

	return pipelineWithInstanceVars.String(), nil
}

func instanceVars(query url.Values) (string, error) {
	var instanceArgs strings.Builder

	if len(query) == 0 {
		return "", nil
	}

	for k, v := range query {
		if !strings.HasPrefix(k, "vars.") {
			continue
		}

		if len(v) > 1 {
			return "", fmt.Errorf("parsing instance variables: expecting ecactly one value for %s, but found %d", k, len(v))
		}

		instanceArgs.WriteString("--instance-var ")
		instanceArgs.WriteString(strings.TrimPrefix(k, "vars."))
		instanceArgs.WriteString("=")
		instanceArgs.WriteString(v[0])

		// TODO Add space if more than one
	}

	return instanceArgs.String(), nil
}

// copied from https://github.com/concourse/concourse/blob/6984e4d30a35f378d31d5897c5a6da2606b62f58/fly/commands/hijack.go/#L239-L252
func parseUrlPath(urlPath string) map[string]string {
	pathWithoutFirstSlash := strings.Replace(urlPath, "/", "", 1)
	urlComponents := strings.Split(pathWithoutFirstSlash, "/")
	urlMap := make(map[string]string)

	for i := 0; i < len(urlComponents)/2; i++ {
		keyIndex := i * 2
		valueIndex := keyIndex + 1
		urlMap[urlComponents[keyIndex]] = urlComponents[valueIndex]
	}

	return urlMap
}
