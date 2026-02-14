package runner

import (
	"os"
	"os/exec"
	"strings"

	"github.com/loheagn/gun/internal/errs"
	"github.com/loheagn/gun/internal/locator"
)

type Invocation struct {
	Dir  string
	Args []string
}

func BuildInvocation(res locator.Resolution, passthrough []string) (Invocation, error) {
	pkgTarget := "."
	dir := res.PackageDir
	useRun := false

	switch res.Mode {
	case locator.ModeProject:
		pkgTarget = "./..."
		dir = res.ModuleRoot
	case locator.ModePkg:
		pkgTarget = "."
	default:
		pkgTarget = "."
		useRun = true
	}

	if useRun && hasRunOverride(passthrough) {
		return Invocation{}, errs.New(errs.CodeUsage, "do not pass -run for this command; gun already selects tests", nil)
	}

	args := []string{"test"}
	if useRun {
		args = append(args, "-run", res.RunPattern)
	}
	args = append(args, passthrough...)
	args = append(args, pkgTarget)

	return Invocation{Dir: dir, Args: args}, nil
}

func Run(inv Invocation) error {
	cmd := exec.Command("go", inv.Args...)
	cmd.Dir = inv.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errs.New(errs.CodeTestFailed, "go test failed", err)
	}
	return nil
}

func hasRunOverride(args []string) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-run" {
			return true
		}
		if strings.HasPrefix(arg, "-run=") {
			return true
		}
	}
	return false
}
