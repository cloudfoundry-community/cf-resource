package in_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/in"
)

var _ = Describe("In", func() {
	var (
		tmpDir   string
		request  in.Request
		response in.Response
	)

	JustBeforeEach(func() {
		binPath, err := gexec.Build("github.com/concourse/cf-resource/in/cmd/in")
		Ω(err).ShouldNot(HaveOccurred())

		tmpDir, err = ioutil.TempDir("", "cf_resource_in")

		stdin := &bytes.Buffer{}
		err = json.NewEncoder(stdin).Encode(request)
		Ω(err).ShouldNot(HaveOccurred())

		cmd := exec.Command(binPath, tmpDir)
		cmd.Stdin = stdin
		cmd.Dir = tmpDir

		session, err := gexec.Start(
			cmd,
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		err = json.Unmarshal(session.Out.Contents(), &response)
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("when a version is given to the executable", func() {
		BeforeEach(func() {
			request = in.Request{
				Source: resource.Source{
					API:           "https://api.run.pivotal.io",
					Username:      "awesome@example.com",
					Password:      "hunter2",
					Organization:  "org",
					Space:         "space",
					SkipCertCheck: true,
				},
				Version: resource.Version{
					Timestamp: time.Now().Add(332 * time.Hour),
				},
			}
		})

		It("outputs that version", func() {
			Ω(response.Version.Timestamp).Should(BeTemporally("~", request.Version.Timestamp, time.Second))
		})
	})

	Context("when a version is not given to the executable", func() {
		BeforeEach(func() {
			request = in.Request{
				Source: resource.Source{
					API:           "https://api.run.pivotal.io",
					Username:      "awesome@example.com",
					Password:      "hunter2",
					Organization:  "org",
					Space:         "space",
					SkipCertCheck: true,
				},
			}
		})

		It("generates a 'fake' current version", func() {
			Ω(response.Version.Timestamp).Should(BeTemporally("~", time.Now(), time.Second))
		})
	})
})
