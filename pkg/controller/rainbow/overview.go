package rainbow

import (
	"context"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/model"
)

func (s *ServerController) Overview(ctx context.Context) (interface{}, error) {
	if SwrClient == nil {
		return nil, nil
	}

	return SwrClient.ShowDomainOverview(&model.ShowDomainOverviewRequest{})
}

func (s *ServerController) Downflow(ctx context.Context) (interface{}, error) {
	if SwrClient == nil {
		return nil, nil
	}

	request := &model.ShowDomainResourceReportsRequest{}
	request.ResourceType = model.GetShowDomainResourceReportsRequestResourceTypeEnum().DOWNFLOW
	request.Frequency = model.GetShowDomainResourceReportsRequestFrequencyEnum().DAILY
	return SwrClient.ShowDomainResourceReports(request)

}
func (s *ServerController) Store(ctx context.Context) (interface{}, error) {
	if SwrClient == nil {
		return nil, nil
	}

	request := &model.ShowDomainResourceReportsRequest{}
	request.ResourceType = model.GetShowDomainResourceReportsRequestResourceTypeEnum().STORE
	request.Frequency = model.GetShowDomainResourceReportsRequestFrequencyEnum().DAILY
	return SwrClient.ShowDomainResourceReports(request)
}
