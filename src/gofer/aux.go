package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func do(cmdline string, params ...interface{}) {
	cmd := fmt.Sprintf(cmdline, params...)
	fmt.Printf("do: %s\n", cmd)
	args := strings.Split(cmd, " ")
	assert(exec.Command(args[0], args[1:]...).Run())
}
