package fork_split_openfeature_provider_go

//go:generate go run go.uber.org/mock/mockgen -package mocks -source=splitClient.go -destination=mocks/mockSplitClient.go

type ISplitClient interface {
	Treatment(key any, feature string, attributes map[string]any) string
}
