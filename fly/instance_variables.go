package fly

import (
	"fmt"
	"net/url"
	"strings"
)

func PipelineWithInstanceVars(pipeline string, query url.Values) (string, error) {
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

func InstanceVars(query url.Values) ([]string, error) {
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
