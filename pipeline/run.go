package pipeline

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func Run(leftCommand, rightCommand *exec.Cmd) error {
	vipe := exec.Command("vipe", "--suffix=yaml")

	reader, writer := io.Pipe()
	vipe.Stdin = reader
	vipe.Stdout = os.Stdout

	go func() {
		defer writer.Close()

		out, err := leftCommand.StdoutPipe()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting output of left command: %s\n", err)
			os.Exit(1)
		}

		err = leftCommand.Start()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running left command: %s\n", err)
			os.Exit(1)
		}

		_, err = io.Copy(writer, out)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error copying output of left command into vipe: %s\n", err)
			os.Exit(1)
		}

		if err := leftCommand.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "Error waiting for left command to complete: %s\n", err)
			os.Exit(1)
		}
	}()

	if err := vipe.Run(); err != nil {
		return fmt.Errorf("running vipe: %w", err)
	}

	return nil
}
