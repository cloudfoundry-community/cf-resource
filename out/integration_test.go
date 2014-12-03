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

	"github.com/concourse/cf-resource/out"
)

var _ = Describe("Out", func() {
	var (
		tmpDir string
		cmd    *exec.Cmd
	)

	BeforeEach(func() {
		binPath, err := gexec.Build("github.com/concourse/cf-resource/out/cmd/out")
		Ω(err).ShouldNot(HaveOccurred())

		tmpDir, err = ioutil.TempDir("", "cf_resource_out")

		assetsPath, err := filepath.Abs("assets")
		Ω(err).ShouldNot(HaveOccurred())

		request := out.Request{
			Source: out.Source{
				API:            "https://api.run.pivotal.io",
				Username:       "awesome@example.com",
				Password:       "hunter2",
				Organization:   "org",
				Space:          "space",
				SkipCertCheck:  true,
				CurrentAppName: "awesome-app",
			},
			Params: out.Params{
				ManifestPath: "project/manifest.yml",
				Path:         "another-project",
			},
		}
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

		gexec.CleanupBuildArtifacts()
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
		Ω(session.Err).Should(gbytes.Say("cf zero-downtime-push awesome-app -f %s -p %s",
			filepath.Join(tmpDir, "project/manifest.yml"),
			filepath.Join(tmpDir, "another-project"),
		))
	})
})
