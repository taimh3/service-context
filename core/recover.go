package core

import (
	"log/slog"
)

func Recover() {
	if r := recover(); r != nil {
		// sctx.GlobalLogger().GetLogger("recovered").Errorln(r)
		slog.Error("%+v \n", r)
	}
}
