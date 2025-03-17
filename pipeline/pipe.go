package pipeline

import (
	"bytes"
	"io"
	"os/exec"
)

// Modified version of https://stackoverflow.com/a/26541826
func Run(stdout io.Writer, commands ...*exec.Cmd) (err error) {
	var stderr bytes.Buffer
	pipes := make([]*io.PipeWriter, len(commands)-1)

	i := 0
	for ; i < len(commands)-1; i++ {
		inPipe, outPipe := io.Pipe()
		commands[i].Stdout = outPipe
		commands[i].Stderr = &stderr
		commands[i+1].Stdin = inPipe
		pipes[i] = outPipe
	}

	commands[i].Stdout = stdout
	commands[i].Stderr = &stderr

	return call(commands, pipes)
}

func call(commands []*exec.Cmd, pipes []*io.PipeWriter) error {
	if commands[0].Process == nil {
		err := commands[0].Start()

		if err != nil {
			return err
		}
	}

	if len(commands) > 1 {
		err := commands[1].Start()

		if err != nil {
			return err
		}

		defer func() {
			if err == nil {
				pipes[0].Close()
				err = call(commands[1:], pipes[1:])
			}
		}()
	}

	return commands[0].Wait()
}
