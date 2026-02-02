package rainbow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/pixiulib/exec"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/timeutil"
)

func (s *AgentController) search(ctx context.Context, date []byte) error {
	var reqMeta types.RemoteMetaRequest
	if err := json.Unmarshal(date, &reqMeta); err != nil {
		klog.Errorf("failed to unmarshal remote meta request %v", err)
		return err
	}

	var (
		result []byte
		err    error
	)
	switch reqMeta.Type {
	case types.SearchTypeRepo:
	case types.SearchTypeTag:
	case types.SearchTypeTagInfo:
	default:
		return fmt.Errorf("unsupported req type %d", reqMeta.Type)
	}

	statusCode, errMessage := 0, ""
	if err != nil {
		statusCode, errMessage = 1, err.Error()
		klog.Errorf("远程搜索失败 %v", err)
	}
	data, err := json.Marshal(types.SearchResult{Result: result, ErrMessage: errMessage, StatusCode: statusCode})
	if err != nil {
		klog.Errorf("序列化查询结果失败 %v", err)
		return fmt.Errorf("序列化查询结果失败 %v", err)
	}

	// 保存 30s
	if _, err := s.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, reqMeta.Uid, data, 30*time.Second)
		pipe.Publish(ctx, fmt.Sprintf("__keyspace@0__:%s", reqMeta.Uid), "set")
		return nil
	}); err != nil {
		klog.Errorf("临时存储失败 %v", err)
		return err
	}

	klog.Infof("搜索(%s)结果已暂存, key(%s)", reqMeta.RepositorySearchRequest.Query, reqMeta.Uid)
	return nil
}

func (s *AgentController) SearchGcrRepositories(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error) {
	klog.Infof("搜索 gcr 镜像 %v", opt.Query)
	// https://gcr.io/v2/google-containers/kibana/tags/list
	// https://gcr.io/v2/google-containers/tags/list
	// https://registry.k8s.io/v2/kube-apiserver/tags/list
	// crane ls registry.k8s.io/kube-apiserver
	if len(opt.Namespace) == 0 {
		opt.Namespace = types.DefaultGCRNamespace
	}

	baseURL := fmt.Sprintf("https://gcr.io/v2/%s/tags/list", opt.Namespace)
	resp, err := DoHttpRequest(baseURL)
	if err != nil {
		return nil, err
	}

	var gcrResult types.SearchGCRResult
	if err = json.Unmarshal(resp, &gcrResult); err != nil {
		return nil, err
	}

	var css []types.CommonSearchRepositoryResult
	for _, child := range gcrResult.Child {
		if strings.Contains(child, opt.Query) { // 服务端过滤
			css = append(css, types.CommonSearchRepositoryResult{
				Name:     fmt.Sprintf("gcr.io/%s/%s", opt.Namespace, child),
				Registry: types.ImageHubGCR,
			})
		}
	}
	return css, nil
}

func (s *AgentController) SearchAllRepositories(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error) {
	// 遍历搜索所有已支持镜像仓库
	searchFuncs := []func(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error){
		s.SearchQuayRepositories,
		//s.SearchGcrRepositories,
	}

	var css []types.CommonSearchRepositoryResult
	diff := len(searchFuncs)

	errCh := make(chan error, diff)
	var wg sync.WaitGroup
	wg.Add(diff)
	for _, sf := range searchFuncs {
		go func(f func(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error)) {
			defer wg.Done()
			res, err := f(ctx, opt)
			if err != nil {
				errCh <- err
			} else {
				css = append(css, res...)
			}
		}(sf)
	}
	wg.Wait()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	default:
	}

	return css, nil
}

func (s *AgentController) runCmd(cmd []string) ([]byte, error) {
	klog.Infof("%s is running", cmd)
	out, err := s.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		klog.Errorf("failed to run %v %v %v", cmd, string(out), err)
		return nil, fmt.Errorf("failed to run %v %v %v", cmd, string(out), err)
	}

	return out, nil
}

func (s *AgentController) SearchQuayTags(ctx context.Context, req types.RemoteTagSearchRequest) ([]byte, error) {
	// https://docs.redhat.com/en/documentation/red_hat_quay/3/html-single/red_hat_quay_api_reference/index#listrepotags
	baseURL := fmt.Sprintf("https://quay.io/api/v1/repository/%s/%s/tag/?page=%d&limit=%d&onlyActiveTags=true", req.Namespace, req.Repository, req.Page, req.PageSize)

	switch req.SearchType {
	case types.AccurateSearch:
		baseURL = fmt.Sprintf("%s&specificTag=%s", baseURL, req.Query)
	default:
		// 默认模糊搜索
		baseURL = fmt.Sprintf("%s&filter_tag_name=like:%s", baseURL, req.Query)
	}
	resp, err := DoHttpRequest(baseURL)
	if err != nil {
		return nil, err
	}
	var quaySearchTagResult types.QuaySearchTagResult
	if err = json.Unmarshal(resp, &quaySearchTagResult); err != nil {
		return nil, err
	}

	// TODO：后续在结构体直接序列化或反序列化
	tagResult := make([]types.CommonTag, 0)
	for _, t := range quaySearchTagResult.Tags {
		var tagSize int64
		if t.Size != nil {
			tagSize = *t.Size
		}

		pt, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", t.LastModified)
		tagResult = append(tagResult, types.CommonTag{
			Name:           t.Name,
			Size:           tagSize,
			LastModified:   timeutil.ToTimeAgo(pt),
			ManifestDigest: t.ManifestDigest,
		})
	}

	pageTags := PaginateCommonTagSlice(tagResult, req.Page, req.PageSize)
	return json.Marshal(types.CommonSearchTagResult{
		Hub:        types.ImageHubQuay,
		Namespace:  req.Namespace,
		Repository: req.Repository,
		Total:      len(pageTags),
		PageSize:   req.PageSize,
		Page:       req.Page,
		TagResult:  pageTags,
	})
}

func (s *AgentController) getManifestForQuay(ctx context.Context, hub, ns, repo, tag string) (*types.Image, error) {
	//fullRepo := fmt.Sprintf("%s/%s/%s:%s", hub, ns, repo, tag)
	return nil, nil
}

// DEPRECATED
func (s *AgentController) makeCommonTagForDockerhub(ctx context.Context, ns string, repo string, tagStr []string) []types.CommonTag {
	t := DockerHubTag{
		Namespace: ns,
		Repo:      repo,
		Tags:      tagStr,
		Result:    map[string]types.SearchDockerhubTagInfoResult{},
	}
	cts := t.GetResults()

	var commonTags []types.CommonTag
	for _, tagS := range tagStr {
		old, ok := cts[tagS]
		if ok {
			commonTags = append(commonTags, types.CommonTag{
				Name:           old.Name,
				Size:           old.FullSize,
				LastModified:   timeutil.ToTimeAgo(old.LastUpdated),
				ManifestDigest: old.Digest,
				//Images:         old.Images,
			})
		} else {
			commonTags = append(commonTags, types.CommonTag{
				Name: old.Name,
			})
		}
	}

	return commonTags
}

func (s *AgentController) SearchGCRagInfo(ctx context.Context, req types.RemoteTagInfoSearchRequest) ([]byte, error) {
	repo := fmt.Sprintf("%s/%s/%s:%s", req.Hub, req.Namespace, req.Repository, req.Tag)
	manifest, err := s.GetImageManifest(repo)
	if err != nil {
		return nil, err
	}
	fmt.Println(manifest)

	return nil, nil
}

func (s *AgentController) GetImageManifest(repo string) (types.ImageManifest, error) {
	cmd := []string{"crane", "manifest", repo}
	d, err := s.runCmd(cmd)
	if err != nil {
		return types.ImageManifest{}, err
	}

	var m types.ImageManifest
	if err = json.Unmarshal(d, &m); err != nil {
		return types.ImageManifest{}, err
	}

	return m, nil
}

func (s *AgentController) SearchGCRTags(ctx context.Context, req types.RemoteTagSearchRequest) ([]byte, error) {
	baseURL := fmt.Sprintf("https://gcr.io/v2/%s/%s/tags/list", req.Namespace, req.Repository)

	resp, err := DoHttpRequest(baseURL)
	if err != nil {
		return nil, err
	}
	var gCRSearchTagResult types.GCRSearchTagResult
	if err = json.Unmarshal(resp, &gCRSearchTagResult); err != nil {
		return nil, err
	}

	var tags []string
	if len(req.Query) != 0 {
		// 搜索场景
		switch req.SearchType {
		case types.AccurateSearch:
			// 精准查询只会有一个命中
			// TODO 优化，直接获取目标 tag 即可
			for _, tg := range gCRSearchTagResult.Tags {
				if tg == req.Query {
					tags = append(tags, tg)
					break
				}
			}
		default:
			for _, tg := range gCRSearchTagResult.Tags {
				if strings.Contains(tg, req.Query) {
					tags = append(tags, tg)
				}
			}
		}
	} else {
		// 全局场景
		tags = gCRSearchTagResult.Tags
	}

	tagResult := s.makeCommonTagForGCR(ctx, req.Namespace, req.Repository, PaginateTagSlice(tags, req.Page, req.PageSize))
	return json.Marshal(types.CommonSearchTagResult{
		Hub:        types.ImageHubGCR,
		Namespace:  req.Namespace,
		Repository: req.Repository,
		Total:      len(tags),
		PageSize:   req.PageSize,
		Page:       req.Page,
		TagResult:  tagResult,
	})
}

type DockerHubTag struct {
	sync.RWMutex

	Namespace string
	Repo      string

	Tags   []string
	Result map[string]types.SearchDockerhubTagInfoResult
}

func (g *DockerHubTag) GetDigest() {
	diff := len(g.Tags)

	var wg sync.WaitGroup
	wg.Add(diff)
	for _, tagS := range g.Tags {
		go func(tag string) {
			defer wg.Done()
			_ = g.getOne(tag)
		}(tagS)
	}
	wg.Wait()
}

func (g *DockerHubTag) getOne(tag string) error {
	reqURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s", g.Namespace, g.Repo, tag)
	resp, err := DoHttpRequest(reqURL)
	if err != nil {
		return err
	}

	var sdt types.SearchDockerhubTagInfoResult
	if err = json.Unmarshal(resp, &sdt); err != nil {
		return err
	}

	g.Lock()
	defer g.Unlock()

	g.Result[tag] = sdt
	return nil
}

func (g *DockerHubTag) GetResults() map[string]types.SearchDockerhubTagInfoResult {
	diff := len(g.Tags)
	var wg sync.WaitGroup
	wg.Add(diff)
	for _, tagS := range g.Tags {
		go func(tag string) {
			defer wg.Done()
			_ = g.getOne(tag)
		}(tagS)
	}
	wg.Wait()

	return g.Result
}

type GCRTag struct {
	sync.RWMutex
	exec exec.Interface

	Namespace string
	Repo      string

	Tags   []string
	Result map[string]string
}

func (g *GCRTag) GetResults() map[string]string {
	diff := len(g.Tags)
	var wg sync.WaitGroup
	wg.Add(diff)
	for _, tagS := range g.Tags {
		go func(tag string) {
			defer wg.Done()
			_ = g.getOne(tag)
		}(tagS)
	}
	wg.Wait()

	return g.Result
}

func (g *GCRTag) getOne(tag string) error {
	cmd := []string{"crane", "digest", fmt.Sprintf("gcr.io/%s/%s:%s", g.Namespace, g.Repo, tag)}
	out, err := util.RunCmd(g.exec, cmd)
	if err != nil {
		return err
	}

	g.Lock()
	defer g.Unlock()

	g.Result[tag] = strings.TrimSpace(string(out))
	return nil
}

func (s *AgentController) SearchQuayRepositories(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error) {
	klog.Infof("搜索 quay.io 镜像 %v", opt.Query)
	// https://docs.projectquay.io/api_quay.html#repo-manage-api
	baseURL := fmt.Sprintf("https://quay.io/api/v1/find/repositories?query=%s&page=%s&page_size=%s", opt.Query, "1", "1")
	resp, err := DoHttpRequest(baseURL)
	if err != nil {
		return nil, err
	}
	var quayResult types.SearchQuayResult
	if err = json.Unmarshal(resp, &quayResult); err != nil {
		return nil, err
	}

	var css []types.CommonSearchRepositoryResult
	for _, rep := range quayResult.Results {
		css = append(css, types.CommonSearchRepositoryResult{
			Name:         fmt.Sprintf("%s/%s", rep.Namespace.Name, rep.Name),
			Registry:     types.ImageHubQuay,
			ShortDesc:    rep.Description,
			Stars:        rep.Stars,
			LastModified: rep.LastModified,
		})
	}
	return css, nil
}

func (s *AgentController) makeCommonTagForGCR(ctx context.Context, ns string, repo string, tagStr []string) []types.CommonTag {
	gcrTag := GCRTag{
		exec:      s.exec,
		Namespace: ns,
		Repo:      repo,
		Tags:      tagStr,
		Result:    map[string]string{},
	}
	resultMap := gcrTag.GetResults()

	cts := make([]types.CommonTag, 0)
	// 排序效果
	for _, tagS := range tagStr {
		digest, ok := resultMap[tagS]
		if ok {
			cts = append(cts, types.CommonTag{
				Name:           tagS,
				ManifestDigest: digest,
			})
		} else {
			cts = append(cts, types.CommonTag{
				Name: tagS,
			})
		}
	}
	return cts
}
