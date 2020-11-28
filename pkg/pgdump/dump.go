package pgdump

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

// Dump is a struct to allow you te generate pg_dumps
type Dump struct {
	postgresHost     string
	postgresPort     int
	postgresUsername string
	postgresDatabase string
	postgresPassword string
}

// NewDump gives a new dump setup for given DB credentials
func NewDump(host string, port int, username, password, database string) *Dump {
	return &Dump{
		postgresHost:     host,
		postgresPort:     port,
		postgresUsername: username,
		postgresDatabase: database,
		postgresPassword: password,
	}
}

func (d *Dump) DumpToDisk(fileName string, options ...string) error {
	options = append(options,
		fmt.Sprintf(`-h%v`, d.postgresHost),
		fmt.Sprintf(`-p%v`, d.postgresPort),
		fmt.Sprintf(`-d%v`, d.postgresDatabase),
		fmt.Sprintf(`-U%v`, d.postgresUsername),
		"-Fc",
		fmt.Sprintf(`-f%v`, fileName))

	cmd := exec.Command("pg_dump", options...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Env = []string{"PGPASSWORD=" + d.postgresPassword}

	err := cmd.Run()
	if err != nil {
		return errors.New(stdout.String() + stderr.String() + err.Error())
	}

	return nil
}

func (d *Dump) RestoreFrom(fileName string, options ...string) error {
	options = append(options,
		fmt.Sprintf(`-h%v`, d.postgresHost),
		fmt.Sprintf(`-p%v`, d.postgresPort),
		fmt.Sprintf(`-d%v`, d.postgresDatabase),
		fmt.Sprintf(`-U%v`, d.postgresUsername),
		"-Fc",
		fmt.Sprintf(`-f%v`, fileName))

	cmd := exec.Command("-g_restore", options...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Env = []string{"PGPASSWORD=" + d.postgresPassword}

	err := cmd.Run()
	if err != nil {
		return errors.New(stdout.String() + stderr.String() + err.Error())
	}

	return nil
}
