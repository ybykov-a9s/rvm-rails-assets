package railsassets

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type PrecompileProcess struct {
	executable Executable
	logger     LogEmitter
}

func NewPrecompileProcess(executable Executable, logger LogEmitter) PrecompileProcess {
	return PrecompileProcess{
		executable: executable,
		logger:     logger,
	}
}

func (p PrecompileProcess) Execute(workingDir string) error {
	buffer := bytes.NewBuffer(nil)

	dbAdapterEnvArg := ""
	env := []string{}
	if val, ok := os.LookupEnv("DB_ADAPTER"); ok {
		dbAdapterEnvArg = "DB_ADAPTER=" + val
		env = append(env, dbAdapterEnvArg)
	}

	args := []string{
		"--login",
		"-c",
		strings.Join(
			[]string{
				"source",
				filepath.Join(os.ExpandEnv("$rvm_path"), "profile.d", "rvm"),
				"&&",
				dbAdapterEnvArg,
				"bundle",
				"exec",
				"rake",
				"assets:precompile",
				"assets:clean",
			},
			" ",
		),
	}

	p.logger.Subprocess("Running 'bash %s'", strings.Join(args, " "))
	err := p.executable.Execute(pexec.Execution{
		Args:   args,
		Env:    env,
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		return fmt.Errorf("failed to execute bundle exec output:\n%s\nerror: %s", buffer.String(), err)
	}

	return nil
}
