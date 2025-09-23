package rainbow

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	"github.com/caoyingjunz/rainbow/pkg/types"
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

func DoHttpRequestWithHeader(url string, header map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest("", url, nil)
	if err != nil {
		return nil, err
	}
	if len(header) != 0 {
		for k, v := range header {
			request.Header.Set(k, v)
		}
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

func PaginateTagSlice(s []string, page int, pageSize int) []string {
	// 验证参数有效性
	if page < 1 || pageSize < 1 || len(s) == 0 {
		return []string{}
	}

	// 计算总页数
	totalItems := len(s)
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	// 检查请求的页码是否超出范围
	if page > totalPages {
		return []string{}
	}

	// 计算起始和结束索引
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	// 确保结束索引不超过数组长度
	if endIndex > totalItems {
		endIndex = totalItems
	}

	return s[startIndex:endIndex]
}

func PaginateCommonTagSlice(s []types.CommonTag, page int, pageSize int) []types.CommonTag {
	// 验证参数有效性
	if page < 1 || pageSize < 1 || len(s) == 0 {
		return []types.CommonTag{}
	}

	// 计算总页数
	totalItems := len(s)
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	// 检查请求的页码是否超出范围
	if page > totalPages {
		return []types.CommonTag{}
	}

	// 计算起始和结束索引
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	// 确保结束索引不超过数组长度
	if endIndex > totalItems {
		endIndex = totalItems
	}

	return s[startIndex:endIndex]
}
