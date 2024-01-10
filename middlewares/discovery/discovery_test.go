package discovery

import (
	"fmt"
	"testing"
	"time"
)

func TestDiscovery(t *testing.T) {
	InitServiceDiscovery([]string{"127.0.0.1:2379"}, []string{"user-svr"})
	sd := GetServiceDiscovery("user-svr")
	fmt.Println("sd", sd)
	time.Sleep(2 * time.Second)
	endpoints := sd.GetHttpEndPoint()
	fmt.Println("endpoints", endpoints)
}
