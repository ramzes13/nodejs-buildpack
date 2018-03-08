package yarn

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Command interface {
	Execute(dir string, stdout io.Writer, stderr io.Writer, program string, args ...string) error
	Run(cmd *exec.Cmd) error
}

type Yarn struct {
	Command Command
	Log     *libbuildpack.Logger
}

func (y *Yarn) Build(buildDir, pkgDir, cacheDir string) error {
	y.Log.Info("Installing node modules (yarn.lock)")

	offline, err := libbuildpack.FileExists(filepath.Join(buildDir, "npm-packages-offline-cache"))
	if err != nil {
		return err
	}
	fmt.Println("Offline:", offline)

	installArgs := []string{"install", "--pure-lockfile", "--ignore-engines", "--cache-folder", filepath.Join(cacheDir, ".cache/yarn")}
	checkArgs := []string{"check"}

	var npmOfflineCache, offlineMirrorPruning string
	if offline {
		y.Log.Info("Found yarn mirror directory %s", npmOfflineCache)
		y.Log.Info("Running yarn in offline mode")

		installArgs = append(installArgs, "--offline")
		checkArgs = append(checkArgs, "--offline")

		npmOfflineCache = filepath.Join(buildDir, "npm-packages-offline-cache")
		offlineMirrorPruning = "false"
	} else {
		y.Log.Info("Running yarn in online mode")
		y.Log.Info("To run yarn in offline mode, see: https://yarnpkg.com/blog/2016/11/24/offline-mirror")

		npmOfflineCache = filepath.Join(cacheDir, "npm-packages-offline-cache")
		offlineMirrorPruning = "true"
	}

	if err := y.Command.Execute(pkgDir, y.Log.Output(), y.Log.Output(), "yarn", "config", "set", "yarn-offline-mirror", npmOfflineCache); err != nil {
		return err
	}
	if err := y.Command.Execute(pkgDir, y.Log.Output(), y.Log.Output(), "yarn", "config", "set", "yarn-offline-mirror-pruning", offlineMirrorPruning); err != nil {
		return err
	}
	if err := y.Command.Execute(pkgDir, y.Log.Output(), y.Log.Output(), "yarn", "config", "set", "cache-folder", filepath.Join(cacheDir, ".cache/yarn")); err != nil {
		return err
	}
	if err := y.Command.Execute(pkgDir, y.Log.Output(), y.Log.Output(), "yarn", "config", "list"); err != nil {
		return err
	}
	if err := y.Command.Execute(pkgDir, y.Log.Output(), y.Log.Output(), "yarn", "config", "current"); err != nil {
		return err
	}

	fmt.Println(installArgs)

	env := os.Environ()
	env = append(env, "YARN_CACHE_FOLDER="+filepath.Join(cacheDir, ".cache/yarn"))
	env = append(env, "npm_config_nodedir="+os.Getenv("NODE_HOME"))
	fmt.Println("npm_config_nodedir=" + os.Getenv("NODE_HOME"))

	cmd := exec.Command("yarn", installArgs...)
	cmd.Dir = pkgDir
	cmd.Stdout = y.Log.Output()
	cmd.Stderr = y.Log.Output()
	cmd.Env = env
	if err := y.Command.Run(cmd); err != nil {
		return err
	}

	if err := y.Command.Execute(pkgDir, ioutil.Discard, os.Stderr, "yarn", checkArgs...); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return err
		}
		y.Log.Warning("yarn.lock is outdated")
	} else {
		y.Log.Info("yarn.lock and package.json match")
	}

	return nil
}
