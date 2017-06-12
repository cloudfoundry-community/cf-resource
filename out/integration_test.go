package out_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/out"
)

var _ = Describe("Out", func() {
	var (
		tmpDir  string
		cmd     *exec.Cmd
		request out.Request
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "cf_resource_out")
		Ω(err).ShouldNot(HaveOccurred())

		err = os.Mkdir(filepath.Join(tmpDir, "project"), 0755)
		Ω(err).ShouldNot(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(tmpDir, "project", "manifest.yml"), []byte{}, 0555)
		Ω(err).ShouldNot(HaveOccurred())

		err = os.Mkdir(filepath.Join(tmpDir, "another-project"), 0555)
		Ω(err).ShouldNot(HaveOccurred())

		request = out.Request{
			Source: resource.Source{
				API:           "https://api.run.pivotal.io",
				Username:      "awesome@example.com",
				Password:      "hunter2",
				Organization:  "org",
				Space:         "space",
				SkipCertCheck: true,
			},
			Params: out.Params{
				ManifestPath:   "project/manifest.yml",
				Path:           "another-project",
				CurrentAppName: "awesome-app",
			},
		}
	})

	JustBeforeEach(func() {
		assetsPath, err := filepath.Abs("assets")
		Ω(err).ShouldNot(HaveOccurred())

		stdin := &bytes.Buffer{}

		err = json.NewEncoder(stdin).Encode(request)
		Ω(err).ShouldNot(HaveOccurred())

		cmd = exec.Command(binPath, tmpDir)
		cmd.Stdin = stdin
		cmd.Dir = tmpDir

		newEnv := []string{}
		for _, envVar := range os.Environ() {
			if strings.HasPrefix(envVar, "PATH=") {
				newEnv = append(newEnv, fmt.Sprintf("PATH=%s:%s", assetsPath, os.Getenv("PATH")))
			} else {
				newEnv = append(newEnv, envVar)
			}
		}

		cmd.Env = newEnv
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("when my manifest and paths do not contain a glob", func() {
		BeforeEach(func() {
			request.Source.CommandEnvironmentVariables = map[string]interface{}{
				"COMMAND_ENV_ONE": "command_env_one",
				"COMMAND_ENV_TWO": "command_env_two",
			}
		})

		It("pushes an application to cloud foundry", func() {
			session, err := gexec.Start(
				cmd,
				GinkgoWriter,
				GinkgoWriter,
			)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			var response out.Response
			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(response.Version.Timestamp).Should(BeTemporally("~", time.Now(), time.Second))

			// shim outputs arguments
			Ω(session.Err).Should(gbytes.Say("cf api https://api.run.pivotal.io --skip-ssl-validation"))
			Ω(session.Err).Should(gbytes.Say("cf auth awesome@example.com hunter2"))
			Ω(session.Err).Should(gbytes.Say("cf target -o org -s space"))
			Ω(session.Err).Should(gbytes.Say("cf zero-downtime-push awesome-app -f %s",
				filepath.Join(tmpDir, "project/manifest.yml"),
			))
			Ω(session.Err).Should(gbytes.Say(filepath.Join(tmpDir, "another-project")))

			// color should be always
			Ω(session.Err).Should(gbytes.Say("CF_COLOR=true"))
			// order is important because `env | sort` as Say fast forwards
			Ω(session.Err).Should(gbytes.Say("COMMAND_ENV_ONE=command_env_one"))
			Ω(session.Err).Should(gbytes.Say("COMMAND_ENV_TWO=command_env_two"))
		})
	})

	Context("when my manifest and file paths contain a glob", func() {
		var tmpFileManifest *os.File
		var tmpFileSearch *os.File

		BeforeEach(func() {
			var err error

			tmpFileManifest, err = ioutil.TempFile(tmpDir, "manifest-some-glob.yml_")
			Ω(err).ShouldNot(HaveOccurred())
			tmpFileSearch, err = ioutil.TempFile(tmpDir, "another-path.jar_")
			Ω(err).ShouldNot(HaveOccurred())

			request.Params.ManifestPath = "manifest-*.yml_*"
			request.Params.Path = "another-path.jar*"
		})

		Context("when one file matches", func() {
			It("pushes an application to cloud foundry", func() {
				session, err := gexec.Start(
					cmd,
					GinkgoWriter,
					GinkgoWriter,
				)
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				var response out.Response
				err = json.Unmarshal(session.Out.Contents(), &response)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response.Version.Timestamp).Should(BeTemporally("~", time.Now(), time.Second))

				// shim outputs arguments
				Ω(session.Err).Should(gbytes.Say("cf api https://api.run.pivotal.io --skip-ssl-validation"))
				Ω(session.Err).Should(gbytes.Say("cf auth awesome@example.com hunter2"))
				Ω(session.Err).Should(gbytes.Say("cf target -o org -s space"))
				Ω(session.Err).Should(gbytes.Say("cf zero-downtime-push awesome-app -f %s -p %s",
					tmpFileManifest.Name(),
					tmpFileSearch.Name(),
				))

				// color should be always
				Ω(session.Err).Should(gbytes.Say("CF_COLOR=true"))
			})
		})

		Context("when no files match the manifest path", func() {
			BeforeEach(func() {
				request.Params.ManifestPath = "nope-*"
			})

			It("returns an error", func() {
				session, err := gexec.Start(
					cmd,
					GinkgoWriter,
					GinkgoWriter,
				)
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))

				errMsg := fmt.Sprintf("error invalid manifest path: found 0 files instead of 1 at path: %s", filepath.Join(tmpDir, `nope-\*`))
				Ω(session.Err).Should(gbytes.Say(errMsg))
			})
		})

		Context("when more then one file matches the manifest path", func() {
			BeforeEach(func() {
				_, err := ioutil.TempFile(tmpDir, "manifest-some-glob.yml_")
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("returns an error", func() {
				session, err := gexec.Start(
					cmd,
					GinkgoWriter,
					GinkgoWriter,
				)
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				errMsg := fmt.Sprintf("error invalid manifest path: found 2 files instead of 1 at path: %s", filepath.Join(tmpDir, `manifest-\*.yml_\*`))
				Ω(session.Err).Should(gbytes.Say(errMsg))
			})
		})

		Context("when no files match the path", func() {
			BeforeEach(func() {
				request.Params.Path = "nope-*"
			})

			It("returns an error", func() {
				session, err := gexec.Start(
					cmd,
					GinkgoWriter,
					GinkgoWriter,
				)
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))

				errMsg := fmt.Sprintf("error invalid path: found 0 files instead of 1 at path: %s", filepath.Join(tmpDir, `nope-\*`))
				Ω(session.Err).Should(gbytes.Say(errMsg))
			})
		})

		Context("when more then one file matches the manifest path", func() {
			BeforeEach(func() {
				_, err := ioutil.TempFile(tmpDir, "another-path.jar_")
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("returns an error", func() {
				session, err := gexec.Start(
					cmd,
					GinkgoWriter,
					GinkgoWriter,
				)
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				errMsg := fmt.Sprintf("error invalid path: found 2 files instead of 1 at path: %s", filepath.Join(tmpDir, `another-path.jar\*`))
				Ω(session.Err).Should(gbytes.Say(errMsg))
			})
		})
	})
})
