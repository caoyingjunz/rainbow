package cmd

import (
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

func ErrorIsNotFound(err error) bool {
	return err.Error() == "record not found"
}

func GetUserInfoByAccessKey(baseURL, accessKey, signature string) (*model.User, error) {
	url := fmt.Sprintf("%s/api/v2/users?access_key=%s", baseURL, accessKey)

	var result UserResult
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodGet).
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{
			"X-ACCESS-KEY":  accessKey,
			"Authorization": signature,
		}).
		Do(&result); err != nil {
		return nil, err
	}
	if result.Code == 200 {
		return &result.Result, nil
	}

	return nil, fmt.Errorf("%s", result.Message)
}

func PrintTable(registries []model.Registry) {
	// 使用 tabwriter 对齐输出
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tID\tCREATED")
	for _, r := range registries {
		created := r.GmtCreate.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%d\t%s\n", r.Name, r.Id, created)
	}
}
