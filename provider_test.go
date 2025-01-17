package split_openfeature_provider_go_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-feature/go-sdk/openfeature"
	. "github.com/splitio/split-openfeature-provider-go"
	"github.com/splitio/split-openfeature-provider-go/mocks"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Provider", func() {
	var (
		mockSplitClient *mocks.MockSplitClient
		subject         *SplitProvider
	)

	BeforeEach(func() {
		mockCtrl := gomock.NewController(GinkgoT())
		mockSplitClient = mocks.NewMockSplitClient(mockCtrl)
		var err error
		subject, err = NewProvider(mockSplitClient)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("Metadata", func() {
		It("should return the correct metadata", func() {
			Ω(subject.Metadata().Name).Should(Equal("Split"))
		})
	})

	Describe("Hooks", func() {
		It("returns an empty list of hooks", func() {
			Ω(subject.Hooks()).Should(BeEmpty())
		})
	})

	Describe("BooleanEvaluation", func() {
		It("should return the default value and error if no targeting key", func() {
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				"foo": uuid.NewString(),
			}

			// act
			result := subject.BooleanEvaluation(context.Background(), feature, true, evalCtx)

			Ω(result.Value).Should(BeTrue())
			Ω(result.ProviderResolutionDetail).Should(Equal(openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTargetingKeyMissingResolutionError("Targeting key is required and missing."),
				Reason:          openfeature.ErrorReason,
				Variant:         "",
			}))
		})

		DescribeTable("split TARGETING_MATCH response",
			func(treatment string, expectedValue bool) {
				key := uuid.NewString()
				feature := uuid.NewString()
				evalCtx := openfeature.FlattenedContext{
					openfeature.TargetingKey: key,
				}
				mockSplitClient.EXPECT().
					Treatment(key, feature, nil).
					Return(treatment)

				// act
				result := subject.BooleanEvaluation(context.Background(), feature, !expectedValue, evalCtx)

				Ω(result).Should(Equal(openfeature.BoolResolutionDetail{
					Value: expectedValue,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						Reason:  openfeature.TargetingMatchReason,
						Variant: treatment,
					},
				}))
			},
			Entry("handles true", "true", true),
			Entry("handles on", "on", true),
			Entry("handles false", "false", false),
			Entry("handles off", "off", false),
		)

		DescribeTable("split CONTROL response",
			func(treatment string) {
				key := uuid.NewString()
				feature := uuid.NewString()
				evalCtx := openfeature.FlattenedContext{
					openfeature.TargetingKey: key,
				}
				mockSplitClient.EXPECT().
					Treatment(key, feature, nil).
					Return(treatment).Times(2)

				// act
				trueDefault := subject.BooleanEvaluation(context.Background(), feature, true, evalCtx)
				falseDefault := subject.BooleanEvaluation(context.Background(), feature, false, evalCtx)

				Ω(trueDefault).Should(Equal(openfeature.BoolResolutionDetail{
					Value: true,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						ResolutionError: openfeature.NewFlagNotFoundResolutionError("Flag not found."),
						Reason:          openfeature.DefaultReason,
						Variant:         treatment,
					},
				}))
				Ω(falseDefault).Should(Equal(openfeature.BoolResolutionDetail{
					Value: false,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						ResolutionError: openfeature.NewFlagNotFoundResolutionError("Flag not found."),
						Reason:          openfeature.DefaultReason,
						Variant:         treatment,
					},
				}))
			},
			Entry("returns default value with empty treatment", ""),
			Entry("returns default value with control treatment", "control"),
		)

		It("returns default value and error if the treatment is not a boolean", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			splitResponse := uuid.NewString()
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(splitResponse)

			// act
			result := subject.BooleanEvaluation(context.Background(), feature, true, evalCtx)

			Ω(result).Should(Equal(openfeature.BoolResolutionDetail{
				Value: true,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					ResolutionError: openfeature.NewParseErrorResolutionError("Error parsing the treatment to the given type."),
					Reason:          openfeature.ErrorReason,
					Variant:         splitResponse,
				},
			}))
		})

		It("passes metadata to the split client", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			attributeValue := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
				"foo":                    attributeValue,
			}
			mockSplitClient.EXPECT().
				Treatment(key, feature, map[string]any{
					"foo": attributeValue,
				}).
				Return("off")

			// act
			result := subject.BooleanEvaluation(context.Background(), feature, true, evalCtx)

			Ω(result).Should(Equal(openfeature.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: "off",
				},
			}))
		})
	})

	Describe("StringEvaluation", func() {
		It("should return the default value and error if no targeting key", func() {
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				"foo": uuid.NewString(),
			}
			defaultValue := uuid.NewString()

			// act
			result := subject.StringEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result.Value).Should(Equal(defaultValue))
			Ω(result.ProviderResolutionDetail).Should(Equal(openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTargetingKeyMissingResolutionError("Targeting key is required and missing."),
				Reason:          openfeature.ErrorReason,
				Variant:         "",
			}))
		})

		It("split TARGETING_MATCH response", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			treatment := uuid.NewString()
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(treatment)

			// act
			result := subject.StringEvaluation(context.Background(), feature, "", evalCtx)

			Ω(result).Should(Equal(openfeature.StringResolutionDetail{
				Value: treatment,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: treatment,
				},
			}))
		})

		DescribeTable("split CONTROL response",
			func(treatment string) {
				key := uuid.NewString()
				feature := uuid.NewString()
				evalCtx := openfeature.FlattenedContext{
					openfeature.TargetingKey: key,
				}
				mockSplitClient.EXPECT().
					Treatment(key, feature, nil).
					Return(treatment)
				defaultValue := uuid.NewString()

				// act
				result := subject.StringEvaluation(context.Background(), feature, defaultValue, evalCtx)

				Ω(result).Should(Equal(openfeature.StringResolutionDetail{
					Value: defaultValue,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						ResolutionError: openfeature.NewFlagNotFoundResolutionError("Flag not found."),
						Reason:          openfeature.DefaultReason,
						Variant:         treatment,
					},
				}))
			},
			Entry("returns default value with empty treatment", ""),
			Entry("returns default value with control treatment", "control"),
		)

		It("passes metadata to the split client", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			attributeValue := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
				"foo":                    attributeValue,
			}
			mockSplitClient.EXPECT().
				Treatment(key, feature, map[string]any{
					"foo": attributeValue,
				}).
				Return("bar")

			// act
			result := subject.StringEvaluation(context.Background(), feature, "", evalCtx)

			Ω(result).Should(Equal(openfeature.StringResolutionDetail{
				Value: "bar",
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: "bar",
				},
			}))
		})
	})

	Describe("FloatEvaluation", func() {
		It("should return the default value and error if no targeting key", func() {
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				"foo": uuid.NewString(),
			}
			const defaultValue = 0.0

			// act
			result := subject.FloatEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result.Value).Should(Equal(defaultValue))
			Ω(result.ProviderResolutionDetail).Should(Equal(openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTargetingKeyMissingResolutionError("Targeting key is required and missing."),
				Reason:          openfeature.ErrorReason,
				Variant:         "",
			}))
		})

		It("split TARGETING_MATCH response", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			const expected = 2.13
			treatment := fmt.Sprintf("%f", expected)
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(treatment)

			// act
			result := subject.FloatEvaluation(context.Background(), feature, 0, evalCtx)

			Ω(result).Should(Equal(openfeature.FloatResolutionDetail{
				Value: expected,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: treatment,
				},
			}))
		})

		DescribeTable("split CONTROL response",
			func(treatment string) {
				key := uuid.NewString()
				feature := uuid.NewString()
				evalCtx := openfeature.FlattenedContext{
					openfeature.TargetingKey: key,
				}
				mockSplitClient.EXPECT().
					Treatment(key, feature, nil).
					Return(treatment)
				const defaultValue = 5.1

				// act
				result := subject.FloatEvaluation(context.Background(), feature, defaultValue, evalCtx)

				Ω(result).Should(Equal(openfeature.FloatResolutionDetail{
					Value: defaultValue,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						ResolutionError: openfeature.NewFlagNotFoundResolutionError("Flag not found."),
						Reason:          openfeature.DefaultReason,
						Variant:         treatment,
					},
				}))
			},
			Entry("returns default value with empty treatment", ""),
			Entry("returns default value with control treatment", "control"),
		)

		It("returns default value and error if the treatment is not a float", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			splitResponse := uuid.NewString()
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(splitResponse)
			const defaultValue = 8.2883

			// act
			result := subject.FloatEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result).Should(Equal(openfeature.FloatResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					ResolutionError: openfeature.NewParseErrorResolutionError("Error parsing the treatment to the given type."),
					Reason:          openfeature.ErrorReason,
					Variant:         splitResponse,
				},
			}))
		})

		It("passes metadata to the split client", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			attributeValue := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
				"foo":                    attributeValue,
			}
			mockSplitClient.EXPECT().
				Treatment(key, feature, map[string]any{
					"foo": attributeValue,
				}).
				Return("821.334")

			// act
			result := subject.FloatEvaluation(context.Background(), feature, 0, evalCtx)

			Ω(result).Should(Equal(openfeature.FloatResolutionDetail{
				Value: 821.334,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: "821.334",
				},
			}))
		})
	})

	Describe("IntEvaluation", func() {
		It("should return the default value and error if no targeting key", func() {
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				"foo": uuid.NewString(),
			}
			const defaultValue = int64(9)

			// act
			result := subject.IntEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result.Value).Should(Equal(defaultValue))
			Ω(result.ProviderResolutionDetail).Should(Equal(openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTargetingKeyMissingResolutionError("Targeting key is required and missing."),
				Reason:          openfeature.ErrorReason,
				Variant:         "",
			}))
		})

		It("split TARGETING_MATCH response", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			const expected = 2
			treatment := fmt.Sprintf("%d", expected)
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(treatment)

			// act
			result := subject.IntEvaluation(context.Background(), feature, 0, evalCtx)

			Ω(result).Should(Equal(openfeature.IntResolutionDetail{
				Value: expected,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: treatment,
				},
			}))
		})

		DescribeTable("split CONTROL response",
			func(treatment string) {
				key := uuid.NewString()
				feature := uuid.NewString()
				evalCtx := openfeature.FlattenedContext{
					openfeature.TargetingKey: key,
				}
				mockSplitClient.EXPECT().
					Treatment(key, feature, nil).
					Return(treatment)
				const defaultValue = 5

				// act
				result := subject.IntEvaluation(context.Background(), feature, defaultValue, evalCtx)

				Ω(result).Should(Equal(openfeature.IntResolutionDetail{
					Value: defaultValue,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						ResolutionError: openfeature.NewFlagNotFoundResolutionError("Flag not found."),
						Reason:          openfeature.DefaultReason,
						Variant:         treatment,
					},
				}))
			},
			Entry("returns default value with empty treatment", ""),
			Entry("returns default value with control treatment", "control"),
		)

		It("returns default value and error if the treatment is not a int", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			splitResponse := uuid.NewString()
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(splitResponse)
			const defaultValue = int64(92)

			// act
			result := subject.IntEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result).Should(Equal(openfeature.IntResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					ResolutionError: openfeature.NewParseErrorResolutionError("Error parsing the treatment to the given type."),
					Reason:          openfeature.ErrorReason,
					Variant:         splitResponse,
				},
			}))
		})

		It("passes metadata to the split client", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			attributeValue := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
				"foo":                    attributeValue,
			}
			mockSplitClient.EXPECT().
				Treatment(key, feature, map[string]any{
					"foo": attributeValue,
				}).
				Return("9923")

			// act
			result := subject.IntEvaluation(context.Background(), feature, 0, evalCtx)

			Ω(result).Should(Equal(openfeature.IntResolutionDetail{
				Value: 9923,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: "9923",
				},
			}))
		})
	})

	Describe("ObjectEvaluation", func() {
		It("should return the default value and error if no targeting key", func() {
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				"foo": uuid.NewString(),
			}
			defaultValue := map[string]any{
				"foo": uuid.NewString(),
				"bar": 293.2,
				"baz": true,
			}

			// act
			result := subject.ObjectEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result.Value).Should(Equal(defaultValue))
			Ω(result.ProviderResolutionDetail).Should(Equal(openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTargetingKeyMissingResolutionError("Targeting key is required and missing."),
				Reason:          openfeature.ErrorReason,
				Variant:         "",
			}))
		})

		It("split TARGETING_MATCH response", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			fooValue := uuid.NewString()
			barValue := 123.456
			treatment := fmt.Sprintf(`{"foo":"%s","bar":%f,"baz":%t}`,
				fooValue, barValue, true)
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(treatment)

			// act
			result := subject.ObjectEvaluation(context.Background(), feature, nil, evalCtx)

			Ω(result).Should(Equal(openfeature.InterfaceResolutionDetail{
				Value: map[string]any{
					"foo": fooValue,
					"bar": barValue,
					"baz": true,
				},
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: treatment,
				},
			}))
		})

		DescribeTable("split CONTROL response",
			func(treatment string) {
				key := uuid.NewString()
				feature := uuid.NewString()
				evalCtx := openfeature.FlattenedContext{
					openfeature.TargetingKey: key,
				}
				mockSplitClient.EXPECT().
					Treatment(key, feature, nil).
					Return(treatment)
				defaultValue := map[string]any{
					"foo": uuid.NewString(),
					"bar": 1979,
					"baz": true,
				}

				// act
				result := subject.ObjectEvaluation(context.Background(), feature, defaultValue, evalCtx)

				Ω(result).Should(Equal(openfeature.InterfaceResolutionDetail{
					Value: defaultValue,
					ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
						ResolutionError: openfeature.NewFlagNotFoundResolutionError("Flag not found."),
						Reason:          openfeature.DefaultReason,
						Variant:         treatment,
					},
				}))
			},
			Entry("returns default value with empty treatment", ""),
			Entry("returns default value with control treatment", "control"),
		)

		It("returns default value and error if the treatment is not json", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
			}
			treatment := uuid.NewString()
			mockSplitClient.EXPECT().
				Treatment(key, feature, nil).
				Return(treatment)
			defaultValue := map[string]any{
				"foo": uuid.NewString(),
				"bar": 1979,
				"baz": true,
			}

			// act
			result := subject.ObjectEvaluation(context.Background(), feature, defaultValue, evalCtx)

			Ω(result).Should(Equal(openfeature.InterfaceResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					ResolutionError: openfeature.NewParseErrorResolutionError("Error parsing the treatment to the given type."),
					Reason:          openfeature.ErrorReason,
					Variant:         treatment,
				},
			}))
		})

		It("passes metadata to the split client", func() {
			key := uuid.NewString()
			feature := uuid.NewString()
			attributeValue := uuid.NewString()
			evalCtx := openfeature.FlattenedContext{
				openfeature.TargetingKey: key,
				"foo":                    attributeValue,
			}
			mockSplitClient.EXPECT().
				Treatment(key, feature, map[string]any{
					"foo": attributeValue,
				}).
				Return(`{"foo":"bar","baz":true}`)

			// act
			result := subject.ObjectEvaluation(context.Background(), feature, nil, evalCtx)

			Ω(result).Should(Equal(openfeature.InterfaceResolutionDetail{
				Value: map[string]any{
					"foo": "bar",
					"baz": true,
				},
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:  openfeature.TargetingMatchReason,
					Variant: `{"foo":"bar","baz":true}`,
				},
			}))
		})
	})

	Describe("NewProviderSimple", Ordered, func() {
		It("successfully creates a new provider", func() {
			Ω(NewProviderSimple("localhost")).ShouldNot(BeNil())
		})

		It("fails with empty api key", func() {
			_, err := NewProviderSimple("")
			Ω(err).Should(HaveOccurred())
		})
	})
})
