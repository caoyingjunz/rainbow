package rainbow

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
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

func DoHttpRequest(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest("", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error resp %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}
