package bin

import (
	"go.uber.org/zap"
)

func init() {
	//traceEnabled = true
	zlog, _ = zap.NewDevelopment()
}
