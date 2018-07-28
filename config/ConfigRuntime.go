package Config

import ()

type ConfigRuntime struct {
	MaxConcurrentNumber int32
	Catalog             string
	Schema              string
	Table               string
	Priority            int32
	S3Region            string
	ParallelNumber      int32

	MaxExecutorNumber int32 //for Agent
}

func NewConfigRuntime() *ConfigRuntime {
	return &ConfigRuntime{
		MaxConcurrentNumber: 10,
		Catalog:             "default",
		Schema:              "default",
		Table:               "default",
		Priority:            0,
		S3Region:            "",
		ParallelNumber:      4,
	}
}
