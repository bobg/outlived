package main

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

func testMode(ctx context.Context, projectID string) error {
	log.Print("starting datastore emulator")
	cmd1 := exec.Command("gcloud", "--project", projectID, "beta", "emulators", "datastore", "start")
	err := cmd1.Start()
	if err != nil {
		return errors.Wrap(err, "starting datastore emulator")
	}

	go func() {
		ch := make(chan struct{})
		go func() {
			defer log.Print("datastore emulator exited")
			defer close(ch)
			err := cmd1.Wait()
			if err != nil {
				log.Printf("datastore emulator: %s", err)
			}
		}()

		select {
		case <-ctx.Done():
			log.Print("sending interrupt to datastore emulator")
			err := cmd1.Process.Signal(os.Interrupt)
			if err != nil {
				log.Fatal(err)
			}

		case <-ch:
		}
	}()

	time.Sleep(5 * time.Second)
	cmd2 := exec.Command("gcloud", "--project", projectID, "beta", "emulators", "datastore", "env-init")
	envLines, err := cmd2.Output()
	if err != nil {
		return errors.Wrap(err, "running env-init command")
	}
	s := bufio.NewScanner(bytes.NewReader(envLines))
	for s.Scan() {
		envLine := s.Text()
		m := envInitRegex.FindStringSubmatch(envLine)
		if len(m) >= 3 {
			err = os.Setenv(m[1], m[2])
			if err != nil {
				return errors.Wrapf(err, "setting env var %s to %s", m[1], m[2])
			}
		}
	}
	return errors.Wrap(s.Err(), "scanning env-init output")
}
