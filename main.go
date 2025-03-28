package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/concourse/concourse/fly/rc"
	"github.com/suhlig/fly-vipe/pipeline"
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

	target, targetName, err := rc.LoadTargetFromURL(fmt.Sprintf("%s://%s", u.Scheme, u.Host), urlMap["teams"], false)

	if err != nil {
		return err
	}

	err = target.Validate()

	if err != nil {
		return err
	}

	pl := urlMap["pipelines"]

	pipelineWithInstanceVars, err := pipelineWithInstanceVars(pl, u.Query())

	if err != nil {
		return err
	}

	instanceVars, err := instanceVars(u.Query())

	if err != nil {
		return err
	}

	p := pipeline.NewPipeline(
		exec.Command("fly",
			fmt.Sprintf("--target=%s", targetName),
			"get-pipeline",
			fmt.Sprintf("--pipeline=%s", pipelineWithInstanceVars),
			fmt.Sprintf("--team=%s", target.Team().Name()),
		),
		exec.Command("vipe",
			"--suffix=yaml",
		),
		exec.Command("fly",
			slices.AppendSeq([]string{ // append another slice to a _literal_ slice of strings
				fmt.Sprintf("--target=%s", targetName),
				"set-pipeline",
				fmt.Sprintf("--pipeline=%s", pl),
				fmt.Sprintf("--team=%s", target.Team().Name()),
				"--config=-",
			}, slices.Values(instanceVars))...,
		),
	)

	// TODO provide an option to just print the pipeline
	fmt.Fprintf(os.Stderr, "Running the following pipeline:\n%s\n", p)

	var stdout, stderr bytes.Buffer

	err = p.Run(&stdout, &stderr)

	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, &stdout)

	return err
}

func pipelineWithInstanceVars(pipeline string, query url.Values) (string, error) {
	var pipelineWithInstanceVars strings.Builder

	pipelineWithInstanceVars.WriteString(pipeline)

	if len(query) > 0 {
		pipelineWithInstanceVars.WriteString("/")

		var instanceArgs []string

		for k, v := range query {
			if !strings.HasPrefix(k, "vars.") {
				continue
			}

			if len(v) > 1 {
				return "", fmt.Errorf("parsing instance variables: expecting ecactly one value for %s, but found %d", k, len(v))
			}

			var pipelineWithInstanceVar strings.Builder

			pipelineWithInstanceVar.WriteString(strings.TrimPrefix(k, "vars."))
			pipelineWithInstanceVar.WriteString(":")
			pipelineWithInstanceVar.WriteString(v[0])

			instanceArgs = append(instanceArgs, pipelineWithInstanceVar.String())
		}

		pipelineWithInstanceVars.WriteString(strings.Join(instanceArgs, ","))
	}

	return pipelineWithInstanceVars.String(), nil
}

func instanceVars(query url.Values) ([]string, error) {
	if len(query) == 0 {
		return nil, nil
	}

	var instanceArgs []string

	for k, v := range query {
		if !strings.HasPrefix(k, "vars.") {
			continue
		}

		if len(v) > 1 {
			return nil, fmt.Errorf("parsing instance variables: expecting ecactly one value for %s, but found %d", k, len(v))
		}

		var instanceArg strings.Builder

		instanceArg.WriteString("--instance-var=")
		instanceArg.WriteString(strings.TrimPrefix(k, "vars."))
		instanceArg.WriteString("=")
		instanceArg.WriteString(v[0])

		instanceArgs = append(instanceArgs, instanceArg.String())
	}

	return instanceArgs, nil
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
