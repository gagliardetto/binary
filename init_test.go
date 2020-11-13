package bin

import "go.uber.org/zap"

func init() {
	//logging.TestingOverride()
	traceEnabled = true
	zlog, _ = zap.NewDevelopment()
}
