package rainbow

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/caoyingjunz/rainbow/pkg/types"
)

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

func ParseImageItem(image string) (string, string, error) {
	parts := strings.Split(image, ":")
	if len(parts) > 2 || len(parts) == 0 {
		return "", "", fmt.Errorf("不合规镜像名称 %s", image)
	}

	path := parts[0]
	tag := "latest"
	if len(parts) == 2 {
		tag = parts[1]
	}

	// 如果镜像是以 docker.io 开关，则去除 docker.io
	if strings.HasPrefix(path, "docker.io/") {
		path = strings.Replace(path, "docker.io/", "", 1)
	}

	return path, tag, nil
}

func WrapNamespace(ns string, user string) string {
	if len(ns) == 0 {
		ns = strings.ToLower(user)
	}
	if ns == defaultNamespace {
		ns = ""
	}

	return ns
}

func ValidateArch(arch string) error {
	parts := strings.Split(arch, "/")
	if len(parts) != 2 && len(parts) != 3 {
		return fmt.Errorf("架构不符合要求，仅支持<操作系统>/<架构> 或 <操作系统>/<架构>/<variant> 格式")
	}

	return nil
}

func ByteSizeSimple(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	base := 1024.0
	exponent := 0
	size := float64(bytes)

	for size >= base && exponent < len(units)-1 {
		size /= base
		exponent++
	}

	return fmt.Sprintf("%.1f %s", size, units[exponent])
}
