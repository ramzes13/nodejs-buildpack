package npm

import (
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Command interface {
	Execute(dir string, stdout io.Writer, stderr io.Writer, program string, args ...string) error
}

type NPM struct {
	Command Command
	Log     *libbuildpack.Logger
}

func (n *NPM) Build(buildDir string) error {
	doBuild, source, err := n.doBuild(buildDir)
	if err != nil {
		return err
	}
	if !doBuild {
		return nil
	}

	n.Log.Info("Installing node modules (%s)", source)
	npmArgs := []string{"install", "--unsafe-perm", "--userconfig", filepath.Join(buildDir, ".npmrc"), "--cache", filepath.Join(buildDir, ".npm")}
	return n.Command.Execute(buildDir, n.Log.Output(), n.Log.Output(), "npm", npmArgs...)
}

func (n *NPM) Rebuild(buildDir string) error {
	doBuild, source, err := n.doBuild(buildDir)
	if err != nil {
		return err
	}
	if !doBuild {
		return nil
	}

	n.Log.Info("Rebuilding any native modules")
	if err := n.Command.Execute(buildDir, n.Log.Output(), n.Log.Output(), "npm", "rebuild", "--nodedir="+os.Getenv("NODE_HOME")); err != nil {
		return err
	}

	n.Log.Info("Installing any new modules (%s)", source)
	npmArgs := []string{"install", "--unsafe-perm", "--userconfig", filepath.Join(buildDir, ".npmrc")}
	return n.Command.Execute(buildDir, n.Log.Output(), n.Log.Output(), "npm", npmArgs...)
}

func (n *NPM) doBuild(buildDir string) (bool, string, error) {
	pkgExists, err := libbuildpack.FileExists(filepath.Join(buildDir, "package.json"))
	if err != nil {
		return false, "", err
	}

	if !pkgExists {
		n.Log.Info("Skipping (no package.json)")
		return false, "", nil
	}

	shrinkwrapExists, err := libbuildpack.FileExists(filepath.Join(buildDir, "npm-shrinkwrap.json"))
	if err != nil {
		return false, "", err
	}

	if shrinkwrapExists {
		return true, "package.json + shrinkwrap", nil
	}
	return true, "package.json", nil
}
