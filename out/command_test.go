package out_test

import (
	"errors"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/cf-resource"
	"github.com/concourse/cf-resource/out"
	"github.com/concourse/cf-resource/out/fakes"
	"io"
	"io/ioutil"
)

var _ = Describe("Out Command", func() {
	var (
		cloudFoundry *fakes.FakePAAS
		request      out.Request
		command      *out.Command
	)

	BeforeEach(func() {
		cloudFoundry = &fakes.FakePAAS{}
		command = out.NewCommand(cloudFoundry)

		request = out.Request{
			Source: resource.Source{
				API:          "https://api.run.pivotal.io",
				Username:     "awesome@example.com",
				Password:     "hunter2",
				Organization: "secret",
				Space:        "volcano-base",
			},
			Params: out.Params{
				ManifestPath: "assets/manifest.yml",
			},
		}
	})

	Describe("running the command", func() {
		It("pushes an application into cloud foundry", func() {
			response, err := command.Run(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.Version.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
			Expect(response.Metadata[0]).To(Equal(
				resource.MetadataPair{
					Name:  "organization",
					Value: "secret",
				},
			))
			Expect(response.Metadata[1]).To(Equal(
				resource.MetadataPair{
					Name:  "space",
					Value: "volcano-base",
				},
			))

			By("logging in")
			Expect(cloudFoundry.LoginCallCount()).To(Equal(1))

			api, username, password, insecure := cloudFoundry.LoginArgsForCall(0)
			Expect(api).To(Equal("https://api.run.pivotal.io"))
			Expect(username).To(Equal("awesome@example.com"))
			Expect(password).To(Equal("hunter2"))
			Expect(insecure).To(Equal(false))

			By("targetting the organization and space")
			Expect(cloudFoundry.TargetCallCount()).To(Equal(1))

			org, space := cloudFoundry.TargetArgsForCall(0)
			Expect(org).To(Equal("secret"))
			Expect(space).To(Equal("volcano-base"))

			By("pushing the app")
			Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))

			manifest, path, currentAppName := cloudFoundry.PushAppArgsForCall(0)
			Expect(manifest).To(Equal("assets/manifest.yml"))
			Expect(path).To(Equal(""))
			Expect(currentAppName).To(Equal(""))
		})

		Describe("handling any errors", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("it all went wrong")
			})

			It("from logging in", func() {
				cloudFoundry.LoginReturns(expectedError)

				_, err := command.Run(request)
				Expect(err).To(MatchError(expectedError))
			})

			It("from targetting an org and space", func() {
				cloudFoundry.TargetReturns(expectedError)

				_, err := command.Run(request)
				Expect(err).To(MatchError(expectedError))
			})

			It("from pushing the application", func() {
				cloudFoundry.PushAppReturns(expectedError)

				_, err := command.Run(request)
				Expect(err).To(MatchError(expectedError))
			})
		})

		Context("setting environment variables provided as params", func() {
			var err error
			var tempFile *os.File

			BeforeEach(func() {
				sourceFile, err := os.Open("assets/manifest.yml")
				Expect(err).NotTo(HaveOccurred())
				defer sourceFile.Close()

				tempFile, err = ioutil.TempFile("assets", "command_test.yml_")
				Expect(err).NotTo(HaveOccurred())
				defer tempFile.Close()

				_, err = io.Copy(tempFile, sourceFile)

				request = out.Request{
					Source: resource.Source{
						API:          "https://api.run.pivotal.io",
						Username:     "awesome@example.com",
						Password:     "hunter2",
						Organization: "secret",
						Space:        "volcano-base",
					},
					Params: out.Params{
						ManifestPath: tempFile.Name(),
						EnvironmentVariables: map[string]string{
							"COMMAND_TEST_A": "command_test_a",
							"COMMAND_TEST_B": "command_test_b",
						},
					},
				}
				_, err = command.Run(request)
			})

			AfterEach(func() {
				os.Remove(tempFile.Name())
			})

			It("does not raise an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes the variables into the manifest", func() {
				manifest, _ := out.NewManifest(request.Params.ManifestPath)

				Expect(manifest.EnvironmentVariables()["COMMAND_TEST_A"]).To(Equal("command_test_a"))
				Expect(manifest.EnvironmentVariables()["COMMAND_TEST_B"]).To(Equal("command_test_b"))
			})
		})

		Context("no environment variables provided", func() {
			It("doesn't set the environment variables", func() {
				manifest, err := out.NewManifest(request.Params.ManifestPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(manifest.EnvironmentVariables()).To(HaveLen(2))
				Expect(manifest.EnvironmentVariables()).To(HaveKeyWithValue("MANIFEST_A", "manifest_a"))
				Expect(manifest.EnvironmentVariables()).To(HaveKeyWithValue("MANIFEST_B", "manifest_b"))
			})
		})

		It("lets people skip the certificate check", func() {
			request = out.Request{
				Source: resource.Source{
					API:           "https://api.run.pivotal.io",
					Username:      "awesome@example.com",
					Password:      "hunter2",
					Organization:  "secret",
					Space:         "volcano-base",
					SkipCertCheck: true,
				},
				Params: out.Params{
					ManifestPath: "a/path/to/a/manifest.yml",
				},
			}

			_, err := command.Run(request)
			Expect(err).NotTo(HaveOccurred())

			By("logging in")
			Expect(cloudFoundry.LoginCallCount()).To(Equal(1))

			_, _, _, insecure := cloudFoundry.LoginArgsForCall(0)
			Expect(insecure).To(Equal(true))
		})

		It("lets people do a zero downtime deploy", func() {
			request = out.Request{
				Source: resource.Source{
					API:          "https://api.run.pivotal.io",
					Username:     "awesome@example.com",
					Password:     "hunter2",
					Organization: "secret",
					Space:        "volcano-base",
				},
				Params: out.Params{
					ManifestPath:   "a/path/to/a/manifest.yml",
					CurrentAppName: "cool-app-name",
				},
			}

			_, err := command.Run(request)
			Expect(err).NotTo(HaveOccurred())

			By("pushing the app")
			Expect(cloudFoundry.PushAppCallCount()).To(Equal(1))

			_, _, currentAppName := cloudFoundry.PushAppArgsForCall(0)
			Expect(currentAppName).To(Equal("cool-app-name"))
		})
	})
})
