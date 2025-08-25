package cmd

import (
	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
)

type config struct {
}

func NewConfig() *config {
	return &config{}
}

func (c *config) ID() string {
	return common.KeyCompConf
}

func (c *config) InitFlags() {
}

func (c *config) Activate(_ sctx.ServiceContext) error {
	return nil
}

func (c *config) Stop() error {
	return nil
}
