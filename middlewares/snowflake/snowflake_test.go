package snowflake

import (
	"testing"
	"time"
)

func TestSnowflake(t *testing.T) {
	Init(time.Now(), 1)
	for i := 0; i < 100; i++ {
		t.Log(GenID())
		t.Log(GenIDString())
	}
}
