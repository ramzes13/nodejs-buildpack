package checksum_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/checksum"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checksum", func() {
	var (
		dir   string
		lines []string
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "checksum")
		Expect(err).To(BeNil())

		Expect(os.MkdirAll(filepath.Join(dir, "a/b"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(dir, "a/b", "file"), []byte("hi"), 0644)).To(Succeed())

		lines = []string{}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	debug := func(format string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}

	Describe("NewChecksum", func() {
		It("Reports the current directory checksum", func() {
			_ = checksum.New(dir, debug)
			Expect(lines).To(Equal([]string{
				"BuildDir Checksum Before Supply: 3e673106d28d587c5c01b3582bf15a50",
			}))
		})
	})

	Describe("After", func() {
		var c *checksum.Checksum
		BeforeEach(func() {
			c = checksum.New(dir, debug)
			time.Sleep(10 * time.Millisecond)
		})
		Context("Directory is unchanged", func() {
			It("Reports the current directory checksum", func() {
				c.After()
				Expect(lines).To(Equal([]string{
					"BuildDir Checksum Before Supply: 3e673106d28d587c5c01b3582bf15a50",
					"BuildDir Checksum After Supply: 3e673106d28d587c5c01b3582bf15a50",
				}))
			})
		})

		Context("a file is changed", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(dir, "a/b", "file"), []byte("bye"), 0644)).To(Succeed())
			})
			It("Reports the current directory checksum", func() {
				c.After()
				Expect(lines).To(Equal([]string{
					"BuildDir Checksum Before Supply: 3e673106d28d587c5c01b3582bf15a50",
					"BuildDir Checksum After Supply: e01956670269656ae69872c0672592ae",
					"Below files changed:",
					"./a/b/file\n",
				}))
			})
		})

		Context("a file is added", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(dir, "a", "file"), []byte("new file"), 0644)).To(Succeed())
			})
			It("Reports the current directory checksum", func() {
				c.After()
				Expect(lines).To(Equal([]string{
					"BuildDir Checksum Before Supply: 3e673106d28d587c5c01b3582bf15a50",
					"BuildDir Checksum After Supply: 9fc7505dc69734c5d40c38a35017e1dc",
					"Below files changed:",
					"./a\n./a/file\n",
				}))
			})
		})
	})
})
