package rainbow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

func (s *AgentController) ProcessSearch(ctx context.Context, req *types.CallSearchRequest) ([]byte, error) {
	switch req.TargetType {
	case types.SearchTypeRepo:
		return s.SearchRepositories(ctx, req)
	case types.SearchTypeTag:
		return s.SearchRepositoryTags(ctx, req)
	case types.SearchTypeTagInfo:
		return s.GetRepositoryTagInfo(ctx, req)
	}

	return nil, fmt.Errorf("unsupported search target type")
}

func (s *AgentController) SearchRepositories(ctx context.Context, req *types.CallSearchRequest) ([]byte, error) {
	var (
		css []types.CommonSearchRepositoryResult
		err error
	)

	switch req.Hub {
	case types.ImageHubDocker:
		css, err = s.SearchDockerHubRepositories(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported hub type %s", req.Hub)
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal(css)
}

func (s *AgentController) SearchRepositoryTags(ctx context.Context, req *types.CallSearchRequest) ([]byte, error) {
	switch req.Hub {
	case types.ImageHubDocker:
		return s.SearchDockerhubTags(ctx, req)
	}

	return nil, fmt.Errorf("unsupported hub type %s", req.Hub)

}

func (s *AgentController) GetRepositoryTagInfo(ctx context.Context, req *types.CallSearchRequest) ([]byte, error) {
	switch req.Hub {
	case types.ImageHubDocker:
		return s.GetDockerhubTagInfo(ctx, req)
	}

	return nil, fmt.Errorf("unsupported hub type %s", req.Hub)
}

func (s *AgentController) SearchDockerHubRepositories(ctx context.Context, req *types.CallSearchRequest) ([]types.CommonSearchRepositoryResult, error) {
	klog.Infof("搜索 dockerhub 镜像 %v", req.Query)

	url := fmt.Sprintf("https://hub.docker.com/v2/search/repositories?query=%s&page=%d&page_size=%d", req.Query, req.Page, req.PageSize)
	var searchResp types.HubSearchResponse
	httpClient := util.HttpClientV2{URL: url}
	err := httpClient.Method(http.MethodGet).
		WithTimeout(30 * time.Second).
		Do(&searchResp)
	if err != nil {
		return nil, err
	}

	var css []types.CommonSearchRepositoryResult
	for _, rep := range searchResp.Results {
		desc := rep.ShortDescription
		css = append(css, types.CommonSearchRepositoryResult{
			Name:       rep.RepoName,
			Registry:   types.ImageHubDocker,
			ShortDesc:  &desc,
			Stars:      rep.StarCount,
			Pull:       rep.PullCount,
			IsOfficial: rep.IsOfficial,
		})
	}

	return css, nil
}

func (s *AgentController) SearchDockerhubTags(ctx context.Context, req *types.CallSearchRequest) ([]byte, error) {
	// https://docs.docker.com/reference/api/registry/latest/#tag/Manifests
	// https://docs.docker.com/reference/api/hub/latest/#tag/repositories/operation/GetRepositoryTag
	// repo=langgenius/dify-api
	// token="$(curl -fsSL "https://auth.docker.io/token?service=registry.docker.io&scope=repository:$repo:pull" | jq --raw-output '.token')"
	// curl -s -H "Authorization: Bearer $token" "https://registry-1.docker.io/v2/$repo/tags/list"
	// curl -H "Authorization: Bearer $token" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" https://registry-1.docker.io/v2/$repo/manifests/latest

	// 获取 token
	repo := fmt.Sprintf("%s/%s", req.Namespace, req.Repository)
	klog.Infof("搜索 dockerhub 镜像(%s) tags", repo)

	var ds types.HubTagResponse
	url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/%s/repositories/%s/tags?page=%d&page_size=%d&name=%s", req.Namespace, req.Repository, req.Page, req.PageSize, req.Query)
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodGet).
		WithTimeout(30 * time.Second).
		Do(&ds); err != nil {
		klog.Errorf("获取镜像tags失败 %v", err)
		return nil, err
	}

	return json.Marshal(types.CommonSearchTagResult{
		Hub:        types.ImageHubDocker,
		Namespace:  req.Namespace,
		Repository: req.Repository,
		Total:      ds.Count,
		PageSize:   req.PageSize,
		Page:       req.Page,
		TagResult:  s.buildCommonTagForDockerhub(ds.Results),
	})
}

func (s *AgentController) GetDockerhubTagInfo(ctx context.Context, req *types.CallSearchRequest) ([]byte, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s/", req.Namespace, req.Repository, req.Tag)

	var searchDockerhubImageResult types.SearchDockerhubTagInfoResult
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodGet).WithTimeout(30 * time.Second).Do(&searchDockerhubImageResult); err != nil {
		return nil, err
	}

	return json.Marshal(types.CommonSearchTagInfoResult{
		Name:     req.Tag,
		FullSize: searchDockerhubImageResult.FullSize,
		Digest:   searchDockerhubImageResult.Digest,
		Images:   searchDockerhubImageResult.Images,
	})
}

func (s *AgentController) buildCommonTagForDockerhub(tags []types.TagResult) []types.CommonTag {
	var cts []types.CommonTag
	for _, t := range tags {
		cts = append(cts, types.CommonTag{
			Name:           t.Name,
			Size:           t.FullSize,
			LastModified:   t.LastUpdated.String(),
			ManifestDigest: t.Digest,
			Images:         t.Images,
		})
	}

	return cts
}
