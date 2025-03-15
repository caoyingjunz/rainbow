package huaweicloud

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	swr "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/region"
)

type HuaweiCloudConfig struct {
	AK       string
	SK       string
	RegionId string
}

func NewHuaweiCloudClient(cfg HuaweiCloudConfig) (*swr.SwrClient, error) {
	auth, err := basic.NewCredentialsBuilder().WithAk(cfg.AK).WithSk(cfg.SK).SafeBuild()
	if err != nil {
		panic(err)
	}

	reg, err := region.SafeValueOf(cfg.RegionId)
	if err != nil {
		return nil, err
	}
	hc, err := swr.SwrClientBuilder().
		WithRegion(reg).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	return swr.NewSwrClient(hc), nil
}
