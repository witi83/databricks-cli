package bundle

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/databricks/cli/libs/git"
	"github.com/databricks/cli/libs/template"
	"github.com/spf13/cobra"
)

var gitUrlPrefixes = []string{
	"https://",
	"git@",
}

func isRepoUrl(url string) bool {
	result := false
	for _, prefix := range gitUrlPrefixes {
		if strings.HasPrefix(url, prefix) {
			result = true
			break
		}
	}
	return result
}

// Computes the repo name from the repo URL. Treats the last non empty word
// when splitting at '/' as the repo name. For example: for url git@github.com:databricks/cli.git
// the name would be "cli.git"
func repoName(url string) string {
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	return parts[len(parts)-1]
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init TEMPLATE_PATH",
		Short: "Initialize Template",
		Args:  cobra.ExactArgs(1),
	}

	var configFile string
	var outputDir string
	var templateDir string
	cmd.Flags().StringVar(&configFile, "config-file", "", "File containing input parameters for template initialization.")
	cmd.Flags().StringVar(&templateDir, "template-dir", "", "Directory within repository that holds the template specification.")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to write the initialized template to.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		templatePath := args[0]
		ctx := cmd.Context()

		if !isRepoUrl(templatePath) {
			// skip downloading the repo because input arg is not a URL. We assume
			// it's a path on the local file system in that case
			return template.Materialize(ctx, configFile, templatePath, outputDir)
		}

		// Download the template in a temporary directory
		tmpDir := os.TempDir()
		templateURL := templatePath
		repoDir := filepath.Join(tmpDir, repoName(templateURL))
		err := os.MkdirAll(repoDir, 0755)
		if err != nil {
			return err
		}
		// TODO: Add automated test that the downloaded git repo is cleaned up.
		err = git.Clone(ctx, templateURL, "", repoDir)
		if err != nil {
			return err
		}
		defer os.RemoveAll(templateDir)
		return template.Materialize(ctx, configFile, filepath.Join(repoDir, templateDir), outputDir)
	}

	return cmd
}