package log

import (
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/constant"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestLog(t *testing.T) {
	Init(WithLogLevel("debug"),
		WithFileName("bitstorm.log"),
		WithMaxSize(100),
		WithMaxBackups(3),
		WithLogPath("./log"),
		WithConsole(true), // 默认不会输出到控制台
	)

	convey.Convey("TestGetLog", t, func() {
		fmt.Println(GetDefaultLogger() == nil, GetDefaultLogger())
	})
	convey.Convey("TestLog", t, func() {
		Infof("bitstorm test %v", "success")
		Errorf("bitstorm test %v", "success")
		Warnf("bitstorm test %v", "success")

		Error("bitstorm test")
		Info("bitstorm test")
		Warn("bitstorm test")

		ctx := context.Background()
		InfoContextf(context.WithValue(ctx, constant.TraceID, "123132321"), "a is %d", 1)
		ErrorContextf(context.WithValue(ctx, constant.TraceID, "123132321"), "a is %d", 1)
		WarnContextf(context.WithValue(ctx, constant.TraceID, "123132321"), "a is %d", 1)
		DebugContextf(context.WithValue(ctx, constant.TraceID, "123132321"), "a is %d", 1)

		//  Fatalf("bitstorm test")

	})

}
