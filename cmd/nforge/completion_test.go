package nforge

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionBash(t *testing.T) {
	cmd := newRootCmdForTest()
	var buf bytes.Buffer
	err := cmd.GenBashCompletionV2(&buf, true)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "complete -o default -F", "bash completion should contain complete command")
	assert.Contains(t, output, "nforge", "bash completion should reference nforge")
}

func TestCompletionZsh(t *testing.T) {
	cmd := newRootCmdForTest()
	var buf bytes.Buffer
	err := cmd.GenZshCompletion(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "compdef nforge", "zsh completion should contain compdef")
	assert.Contains(t, output, "nforge", "zsh completion should reference nforge")
}

func TestCompletionPowerShell(t *testing.T) {
	cmd := newRootCmdForTest()
	var buf bytes.Buffer
	err := cmd.GenPowerShellCompletion(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "powershell completion for nforge", "powershell completion should contain header")
	assert.Contains(t, output, "nforge", "powershell completion should reference nforge")
}

func TestSubcommandCompletion(t *testing.T) {
	cmd := newRootCmdForTest()
	subCommands := cmd.Commands()
	subCmdNames := make([]string, 0, len(subCommands))
	for _, sc := range subCommands {
		if sc.Name() == "__complete" || sc.Name() == "help" {
			continue // skip internal commands
		}
		subCmdNames = append(subCmdNames, sc.Name())
	}
	expected := []string{"serve", "run", "new", "config", "skill", "session", "doctor", "graph"}
	for _, e := range expected {
		assert.Contains(t, subCmdNames, e, "subcommand %s should be registered", e)
	}
}

func TestCompletionCommandAvailable(t *testing.T) {
	cmd := newRootCmdForTest()
	found := false
	for _, sc := range cmd.Commands() {
		if sc.Name() == "completion" {
			found = true
			break
		}
	}
	assert.True(t, found, "completion subcommand should be available")
}

func newRootCmdForTest() *cobra.Command {
	rootCmd.InitDefaultCompletionCmd()
	return rootCmd
}

