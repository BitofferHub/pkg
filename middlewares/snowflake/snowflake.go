package snowflake

import (
	"github.com/bwmarrin/snowflake"
	"time"
)

var (
	node *snowflake.Node
)

func Init(startTime time.Time, machineID int64) {
	startTimeStr := startTime.Format("2006-01-02 00:00:00")
	var st time.Time
	// 时间为UTC时间，比中国慢8个小时
	st, err := time.Parse("2006-01-02 00:00:00", startTimeStr)
	if err != nil {
		panic(err)
	}
	snowflake.Epoch = st.UnixNano() / 1000000 // 毫秒
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		panic(err)
	}
	return
}

func GenID() int64 {
	return node.Generate().Int64()
}

func GenIDString() string {
	return node.Generate().String()
}
