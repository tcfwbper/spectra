package spectra

import (
	"io/fs"

	"github.com/tcfwbper/spectra/internal/builtin"
)

var builtinWorkflowsFS fs.FS = builtin.Workflows

var builtinAgentsFS fs.FS = builtin.Agents

var builtinSpecFilesFS fs.FS = builtin.SpecFiles
