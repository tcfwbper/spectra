package spectra

// scaffolds_test.go provides stub production-surface placeholders so that test files
// compile before the real production implementation lands. Each placeholder names the
// production symbol it stands in for. Remove this file once the real implementations exist.
//
// Missing production symbols:
//   - NewDirectoryCreator → internal/cmd/spectra/directory_creator.go
//   - NewGitignoreEnsurer → internal/cmd/spectra/gitignore_ensurer.go
//   - NewBuiltinResourceCopier → internal/cmd/spectra/builtin_resource_copier.go

import "io/fs"

// --- DirectoryCreator stub ---

// DirectoryCreator is a test-only stub for the production DirectoryCreator type.
// Remove once internal/cmd/spectra/directory_creator.go is implemented.
type DirectoryCreator struct{}

// NewDirectoryCreator is a test-only stub constructor.
func NewDirectoryCreator() *DirectoryCreator {
	return &DirectoryCreator{}
}

// CreateAll is a test-only stub. The real implementation will create project directories.
func (d *DirectoryCreator) CreateAll(projectRoot string) error {
	panic("DirectoryCreator.CreateAll: stub — production implementation required")
}

// --- GitignoreEnsurer stub ---

// GitignoreEnsurer is a test-only stub for the production GitignoreEnsurer type.
// Remove once internal/cmd/spectra/gitignore_ensurer.go is implemented.
type GitignoreEnsurer struct{}

// NewGitignoreEnsurer is a test-only stub constructor.
func NewGitignoreEnsurer() *GitignoreEnsurer {
	return &GitignoreEnsurer{}
}

// Ensure is a test-only stub. The real implementation will ensure .gitignore has .spectra entry.
func (g *GitignoreEnsurer) Ensure(projectRoot string) error {
	panic("GitignoreEnsurer.Ensure: stub — production implementation required")
}

// --- BuiltinResourceCopier stub ---

// StorageLayoutInterface defines the interface that BuiltinResourceCopier depends on
// for path composition. This mirrors what the production code will require.
type StorageLayoutInterface interface {
	GetWorkflowPath(projectRoot, name string) string
	GetAgentPath(projectRoot, name string) string
}

// BuiltinResourceCopier is a test-only stub for the production BuiltinResourceCopier type.
// Remove once internal/cmd/spectra/builtin_resource_copier.go is implemented.
type BuiltinResourceCopier struct {
	workflowsFS fs.FS
	agentsFS    fs.FS
	specFilesFS fs.FS
	layout      StorageLayoutInterface
}

// NewBuiltinResourceCopier is a test-only stub constructor.
func NewBuiltinResourceCopier(workflowsFS, agentsFS, specFilesFS fs.FS, layout StorageLayoutInterface) *BuiltinResourceCopier {
	return &BuiltinResourceCopier{
		workflowsFS: workflowsFS,
		agentsFS:    agentsFS,
		specFilesFS: specFilesFS,
		layout:      layout,
	}
}

// CopyWorkflows is a test-only stub.
func (b *BuiltinResourceCopier) CopyWorkflows(projectRoot string) ([]string, error) {
	panic("BuiltinResourceCopier.CopyWorkflows: stub — production implementation required")
}

// CopyAgents is a test-only stub.
func (b *BuiltinResourceCopier) CopyAgents(projectRoot string) ([]string, error) {
	panic("BuiltinResourceCopier.CopyAgents: stub — production implementation required")
}

// CopySpecFiles is a test-only stub.
func (b *BuiltinResourceCopier) CopySpecFiles(projectRoot string) ([]string, error) {
	panic("BuiltinResourceCopier.CopySpecFiles: stub — production implementation required")
}
