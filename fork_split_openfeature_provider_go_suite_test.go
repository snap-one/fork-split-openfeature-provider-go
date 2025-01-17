package split_openfeature_provider_go_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSplitOpenFeatureProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Split Open Feature Provider Suite")
}
