package cmd

import (
	"fmt"
	"net/http"
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
