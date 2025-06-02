package rainbow

import (
	"math/rand"
	"time"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
)

func GetRpcClient(clientId string, m map[string]pb.Tunnel_ConnectServer) pb.Tunnel_ConnectServer {
	if m == nil || len(m) == 0 {
		return nil
	}

	// 指定
	if len(clientId) != 0 {
		return m[clientId]
	}

	// 随机
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	rand.Seed(time.Now().UnixNano())
	return m[keys[rand.Intn(len(keys))]]
}
