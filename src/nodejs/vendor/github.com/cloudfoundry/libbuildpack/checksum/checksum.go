package checksum

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Checksum struct {
	dir           string
	debug         func(format string, args ...interface{})
	timestampFile string
}

func New(dir string, debug func(format string, args ...interface{})) *Checksum {
	c := &Checksum{dir: dir, debug: debug}

	if f, err := ioutil.TempFile("", "checksum"); err == nil {
		f.Close()
		c.timestampFile = f.Name()
	}

	if sum, err := c.calc(); err == nil {
		c.debug("BuildDir Checksum Before Supply: %s", sum)
	}

	return c
}

func (c *Checksum) After() {
	if sum, err := c.calc(); err == nil {
		c.debug("BuildDir Checksum After Supply: %s", sum)
	}

	if c.timestampFile != "" {
		if filesChanged, err := (&libbuildpack.Command{}).Output(c.dir, "find", ".", "-newer", c.timestampFile, "-not", "-path", "./.cloudfoundry/*", "-not", "-path", "./.cloudfoundry"); err == nil && filesChanged != "" {
			c.debug("Below files changed:")
			c.debug(filesChanged)
		}
	}
}

func (c *Checksum) calc() (string, error) {
	h := md5.New()
	err := filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			relpath, err := filepath.Rel(c.dir, path)
			if strings.HasPrefix(relpath, ".cloudfoundry/") {
				return nil
			}
			if err != nil {
				return err
			}
			if _, err := io.WriteString(h, relpath); err != nil {
				return err
			}
			if f, err := os.Open(path); err != nil {
				return err
			} else {
				if _, err := io.Copy(h, f); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
