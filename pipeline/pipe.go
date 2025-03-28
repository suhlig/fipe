package pipeline

import (
	"io"
	"iter"
	"os/exec"
	"slices"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

type Pipeline struct {
	commands []*exec.Cmd
}

func NewPipeline(commands ...*exec.Cmd) *Pipeline {
	return &Pipeline{commands: commands}
}

// Modified version of https://stackoverflow.com/a/26541826
func (p Pipeline) String() string {
	return strings.Join(
		slices.Collect(mapFunc(slices.Values(p.commands), func(c *exec.Cmd) string {
			return strings.Join(
				slices.Collect(mapFunc(slices.Values(c.Args), func(a string) string { return shellescape.Quote(a) })),
				" ",
			)
		})),
		" | ",
	)
}

// Modified version of https://stackoverflow.com/a/26541826
func (p *Pipeline) Run(stdout, stderr io.Writer) error {
	pipes := make([]*io.PipeWriter, len(p.commands)-1)

	i := 0

	for ; i < len(p.commands)-1; i++ {
		inPipe, outPipe := io.Pipe()
		p.commands[i].Stdout = outPipe
		p.commands[i].Stderr = stderr
		p.commands[i+1].Stdin = inPipe
		pipes[i] = outPipe
	}

	p.commands[i].Stdout = stdout
	p.commands[i].Stderr = stderr

	return call(p.commands, pipes)
}

func call(commands []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if commands[0].Process == nil {
		err = commands[0].Start()

		if err != nil {
			return err
		}
	}

	if len(commands) > 1 {
		err = commands[1].Start()

		if err != nil {
			return err
		}

		defer func() {
			if err == nil {
				pipes[0].Close()
				err = call(commands[1:], pipes[1:])
			} else {
				_ = commands[1].Wait()
			}
		}()
	}

	return commands[0].Wait()
}

func mapFunc[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for a := range seq {
			if !yield(f(a)) {
				return
			}
		}
	}
}
