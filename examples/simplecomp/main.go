package main

import (
	"log"

	sctx "github.com/taimaifika/service-context"
)

type CanGetValue interface {
	GetValue() string
}

func main() {
	const compId = "foo"

	serviceCtx := sctx.NewServiceContext(
		sctx.WithName("simple-component"),
		sctx.WithComponent(NewSimpleComponent(compId)),
	)

	if err := serviceCtx.Load(); err != nil {
		log.Fatal(err)
	}

	comp := serviceCtx.MustGet(compId).(CanGetValue)

	log.Println(comp.GetValue())

	_ = serviceCtx.Stop()
}
