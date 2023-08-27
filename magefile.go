//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var goexe = "go"

func init() {
	if exe := os.Getenv("GOEXE"); exe != "" {
		goexe = exe
	}

	os.Setenv("GO111MODULE", "on")
}

func Deps() error {
	error := sh.Run(goexe, "mod", "download")
	if error != nil {
		return error
	}

	error = sh.Run(
		goexe,
		"install",
		"honnef.co/go/tools/cmd/staticcheck@2023.1.5",
	)
	if error != nil {
		return error
	}

	return sh.Run(goexe, "install", "github.com/segmentio/golines@v0.11.0")
}

func All() {
	mg.SerialDeps(Build, Test, Coverage, Bench, Fmt, Lint)
}

func Build() error {
	return sh.Run(goexe, "build", "./...")
}

func Test() error {
	return sh.Run(goexe, "test", "./...")
}

func Coverage() error {
	return sh.Run(goexe, "test", "-cover", "./...")
}

func Bench() error {
	return sh.Run(goexe, "test", "-bench=.", "./...")
}

func Fmt() error {
	return sh.Run("golines", "-m", "80", "-w", ".")
}

func Fmtcheck() error {
	s, err := sh.Output("golines", "-m", "80", "--dry-run", ".")

	if err != nil {
		fmt.Print("ERROR: running golines -m 80 --dry-run .")
	}

	if s != "" {
		return errors.New("improperly formatted go files")
	}
	return nil
}

func Lint() error {
	return sh.Run("staticcheck", "./...")
}
