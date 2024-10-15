package core

import sctx "github.com/taimaifika/service-context"

func Recover() {
	if r := recover(); r != nil {
		sctx.GlobalLogger().GetLogger("recovered").Errorln(r)
	}
}
