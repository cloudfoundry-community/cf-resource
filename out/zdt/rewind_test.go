package zdt_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/cf-resource/out/zdt"
)

var _ = Describe("Rewind", func() {
	It("runs through all actions if they're all successful", func() {
		firstRun := false
		secondRun := false

		actions := zdt.Actions{
			Actions: []zdt.Action{
				{
					Forward: func() error {
						firstRun = true
						return nil
					},
				},
				{
					Forward: func() error {
						secondRun = true
						return nil
					},
				},
			},
		}

		err := actions.Execute()
		Expect(err).ToNot(HaveOccurred())

		Expect(firstRun).To(BeTrue())
		Expect(secondRun).To(BeTrue())
	})

	It("stops and runs the zdt of an action if it fails", func() {
		firstRun := false
		secondRun := false
		secondReverseRun := false
		thirdRun := false

		actions := zdt.Actions{
			Actions: []zdt.Action{
				{
					Forward: func() error {
						firstRun = true
						return nil
					},
				},
				{
					Forward: func() error {
						secondRun = true
						return errors.New("disaster")
					},
					ReversePrevious: func() error {
						secondReverseRun = true
						return nil
					},
				},
				{
					Forward: func() error {
						thirdRun = true
						return nil
					},
				},
			},
		}

		err := actions.Execute()
		Expect(err).To(MatchError("disaster"))

		Expect(firstRun).To(BeTrue())
		Expect(secondRun).To(BeTrue())
		Expect(secondReverseRun).To(BeTrue())
		Expect(thirdRun).To(BeFalse())
	})

	It("gives up if the zdt action fails", func() {
		firstRun := false
		secondRun := false
		secondReverseRun := false
		thirdRun := false

		actions := zdt.Actions{
			Actions: []zdt.Action{
				{
					Forward: func() error {
						firstRun = true
						return nil
					},
				},
				{
					Forward: func() error {
						secondRun = true
						return errors.New("disaster")
					},
					ReversePrevious: func() error {
						secondReverseRun = true
						return errors.New("another disaster")
					},
				},
				{
					Forward: func() error {
						thirdRun = true
						return nil
					},
				},
			},
			RewindFailureMessage: "uh oh",
		}

		err := actions.Execute()
		Expect(err).To(MatchError("uh oh: another disaster"))

		Expect(firstRun).To(BeTrue())
		Expect(secondRun).To(BeTrue())
		Expect(secondReverseRun).To(BeTrue())
		Expect(thirdRun).To(BeFalse())
	})

	It("just returns the error if a zdt fails with no reverse message", func() {
		firstRun := false
		secondRun := false
		secondReverseRun := false
		thirdRun := false

		actions := zdt.Actions{
			Actions: []zdt.Action{
				{
					Forward: func() error {
						firstRun = true
						return nil
					},
				},
				{
					Forward: func() error {
						secondRun = true
						return errors.New("disaster")
					},
					ReversePrevious: func() error {
						secondReverseRun = true
						return errors.New("another disaster")
					},
				},
				{
					Forward: func() error {
						thirdRun = true
						return nil
					},
				},
			},
		}

		err := actions.Execute()
		Expect(err).To(MatchError("another disaster"))

		Expect(firstRun).To(BeTrue())
		Expect(secondRun).To(BeTrue())
		Expect(secondReverseRun).To(BeTrue())
		Expect(thirdRun).To(BeFalse())
	})
})
