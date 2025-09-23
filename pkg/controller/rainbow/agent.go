package rainbow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/caoyingjunz/pixiulib/exec"
	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/pixiulib/strutil"
	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
)

type AgentGetter interface {
	Agent() Interface
}
type Interface interface {
	Run(ctx context.Context, workers int) error
	Search(ctx context.Context, date []byte) error
}

type AgentController struct {
	factory     db.ShareDaoFactory
	cfg         rainbowconfig.Config
	redisClient *redis.Client

	queue workqueue.RateLimitingInterface
	exec  exec.Interface

	name     string
	callback string
	baseDir  string
}

func NewAgent(f db.ShareDaoFactory, cfg rainbowconfig.Config, redisClient *redis.Client) *AgentController {
	return &AgentController{
		factory:     f,
		cfg:         cfg,
		redisClient: redisClient,
		name:        cfg.Agent.Name,
		baseDir:     cfg.Agent.DataDir,
		callback:    cfg.Plugin.Callback,
		queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rainbow-agent"),
		exec:        exec.New(),
	}
}

func (s *AgentController) Search(ctx context.Context, date []byte) error {
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
		result, err = s.SearchRepositories(ctx, reqMeta.RepositorySearchRequest)
	case types.SearchTypeTag:
		result, err = s.SearchTags(ctx, reqMeta.TagSearchRequest)
	case types.SearchTypeTagInfo:
		result, err = s.SearchTagInfo(ctx, reqMeta.TagInfoSearchRequest)
	case 4:
		result, err = s.SyncKubernetesTags(ctx, reqMeta.KubernetesTagRequest)
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

func (s *AgentController) SearchRepositories(ctx context.Context, req types.RemoteSearchRequest) ([]byte, error) {
	var (
		css []types.CommonSearchRepositoryResult
		err error
	)

	switch req.Hub {
	case types.ImageHubDocker:
		css, err = s.SearchDockerhubRepositories(ctx, req)
	case types.ImageHubQuay:
		css, err = s.SearchQuayRepositories(ctx, req)
	case types.ImageHubGCR:
		css, err = s.SearchGcrRepositories(ctx, req)
	case types.ImageHubAll:
		css, err = s.SearchAllRepositories(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported hub type %s", req.Hub)
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal(css)
}

func (s *AgentController) SearchDockerhubRepositories(ctx context.Context, req types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error) {
	klog.Infof("搜索 dockerhub 镜像 %v", req.Query)
	url := fmt.Sprintf("https://hub.docker.com/v2/search/repositories?query=%s&page=%s&page_size=%s", req.Query, req.Page, req.PageSize)
	resp, err := DoHttpRequest(url)
	if err != nil {
		return nil, err
	}

	var searchResp types.HubSearchResponse
	if err = json.Unmarshal(resp, &searchResp); err != nil {
		klog.Errorf("序列化 dockerhub 响应失败: %v", err)
		return nil, err
	}
	var css []types.CommonSearchRepositoryResult
	for _, rep := range searchResp.Results {
		desc := rep.ShortDescription
		css = append(css, types.CommonSearchRepositoryResult{
			Name:      rep.RepoName,
			Registry:  types.ImageHubDocker,
			ShortDesc: &desc,
			Stars:     rep.StarCount,
			Pull:      rep.PullCount,
		})
	}

	return css, nil
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
			Name:         fmt.Sprintf("quay.io/%s/%s", rep.Namespace.Name, rep.Name),
			Registry:     types.ImageHubQuay,
			ShortDesc:    rep.Description,
			Stars:        rep.Stars,
			LastModified: rep.LastModified,
		})
	}
	return css, nil
}

func (s *AgentController) SearchAllRepositories(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error) {
	// 遍历搜索所有已支持镜像仓库
	searchFuncs := []func(ctx context.Context, opt types.RemoteSearchRequest) ([]types.CommonSearchRepositoryResult, error){
		s.SearchQuayRepositories,
		s.SearchDockerhubRepositories,
		s.SearchGcrRepositories,
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
		tagResult = append(tagResult, types.CommonTag{
			Name:           t.Name,
			Size:           tagSize,
			LastModified:   t.LastModified,
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

func (s *AgentController) SearchDockerhubTags(ctx context.Context, req types.RemoteTagSearchRequest) ([]byte, error) {
	// https://docs.docker.com/reference/api/registry/latest/#tag/Manifests
	// https://docs.docker.com/reference/api/hub/latest/#tag/repositories/operation/GetRepositoryTag
	// repo=langgenius/dify-api
	// token="$(curl -fsSL "https://auth.docker.io/token?service=registry.docker.io&scope=repository:$repo:pull" | jq --raw-output '.token')"
	// curl -s -H "Authorization: Bearer $token" "https://registry-1.docker.io/v2/$repo/tags/list"
	// curl -H "Authorization: Bearer $token" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" https://registry-1.docker.io/v2/$repo/manifests/latest
	repo := fmt.Sprintf("%s/%s", req.Namespace, req.Repository)
	tokenResp, err := DoHttpRequest(fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo))
	if err != nil {
		return nil, err
	}
	var t types.DockerToken
	if err = json.Unmarshal(tokenResp, &t); err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/tags/list", repo)
	resp, err := DoHttpRequestWithHeader(baseURL, map[string]string{"Authorization": fmt.Sprintf("Bearer %s", t.Token)})
	if err != nil {
		return nil, err
	}
	var ds types.DockerhubSearchTagResult
	if err = json.Unmarshal(resp, &ds); err != nil {
		return nil, err
	}

	var tags []string
	if len(req.Query) != 0 {
		// 搜索场景
		switch req.SearchType {
		case types.AccurateSearch:
			// 精准查询只会有一个命中
			// TODO 优化，直接获取目标 tag 即可
			for _, tg := range ds.Tags {
				if tg == req.Query {
					tags = append(tags, tg)
					break
				}
			}
		default:
			for _, tg := range ds.Tags {
				if strings.Contains(tg, req.Query) {
					tags = append(tags, tg)
				}
			}
		}
	} else {
		// 全局场景
		tags = ds.Tags
	}
	commonTags := s.makeCommonTagForDockerhub(ctx, req.Namespace, req.Repository, PaginateTagSlice(tags, req.Page, req.PageSize))
	return json.Marshal(types.CommonSearchTagResult{
		Hub:        types.ImageHubDocker,
		Namespace:  req.Namespace,
		Repository: req.Repository,
		Total:      len(tags),
		PageSize:   req.PageSize,
		Page:       req.Page,
		TagResult:  commonTags,
	})
}

func (s *AgentController) getDockerhubTag(ns, repo, tag string) (types.SearchDockerhubTagInfoResult, error) {
	reqURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s", ns, repo, tag)
	resp, err := DoHttpRequest(reqURL)
	if err != nil {
		return types.SearchDockerhubTagInfoResult{}, err
	}

	var sdt types.SearchDockerhubTagInfoResult
	if err = json.Unmarshal(resp, &sdt); err != nil {
		return types.SearchDockerhubTagInfoResult{}, err
	}
	return sdt, nil
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
				LastModified:   old.LastUpdated.String(),
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

func (s *AgentController) SearchTags(ctx context.Context, req types.RemoteTagSearchRequest) ([]byte, error) {
	switch req.Hub {
	case types.ImageHubQuay:
		return s.SearchQuayTags(ctx, req)
	case types.ImageHubDocker:
		return s.SearchDockerhubTags(ctx, req)
	case types.ImageHubGCR:
		return s.SearchGCRTags(ctx, req)
	}

	return nil, fmt.Errorf("unsupported hub type %s", req.Hub)
	//switch cfg.ImageFrom {
	//case types.ImageHubDocker:
	//	var tagResults []types.TagResult
	//
	//	baseURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags", req.Namespace, req.Repository)
	//	page := cfg.Page
	//
	//	// TODO: 后续优化
	//	// 1. 架构和策略都不限，直接获取
	//	if len(cfg.Arch) == 0 && cfg.Policy == ".*" {
	//		reqURL := fmt.Sprintf("%s?page_size=%s&page=%s", baseURL, fmt.Sprintf("%d", cfg.Size), fmt.Sprintf("%d", page))
	//		val, err := DoHttpRequest(reqURL)
	//		if err != nil {
	//			return nil, err
	//		}
	//		var tagResp types.HubTagResponse
	//		if err = json.Unmarshal(val, &tagResp); err != nil {
	//			return nil, err
	//		}
	//		tagResults = append(tagResults, tagResp.Results...)
	//	} else {
	//		// 2. 架构和策略至少有一个限制，均需要递归查询
	//		re, err := regexp.Compile(util.ToRegexp(cfg.Policy))
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		for {
	//			if len(tagResults) >= cfg.Size {
	//				break
	//			}
	//			reqURL := fmt.Sprintf("%s?page_size=%s&page=%s", baseURL, "100", fmt.Sprintf("%d", page))
	//			klog.Infof("开始调用 %s 获取镜像tag", reqURL)
	//			val, err := DoHttpRequest(reqURL)
	//			if err != nil {
	//				klog.Errorf("url(%s)请求失败 %v", reqURL, err)
	//				return nil, err
	//			}
	//			var tagResp types.HubTagResponse
	//			if err = json.Unmarshal(val, &tagResp); err != nil {
	//				klog.Errorf("序列化 tag 失败 %v", err)
	//				return nil, err
	//			}
	//
	//			for _, tag := range tagResp.Results {
	//				if len(tagResults) >= cfg.Size {
	//					break
	//				}
	//
	//				if cfg.Policy != ".*" {
	//					// 过滤 policy, 不符合 policy 则忽略
	//					if !re.MatchString(tag.Name) {
	//						continue
	//					}
	//				}
	//				if len(cfg.Arch) != 0 {
	//					newImage := make([]types.Image, 0)
	//					for _, image := range tag.Images {
	//						if image.Architecture == cfg.Arch {
	//							newImage = append(newImage, image)
	//						}
	//					}
	//					tag.Images = newImage // 去除不符合要求的架构镜像
	//				}
	//				tagResults = append(tagResults, tag)
	//			}
	//			page++
	//		}
	//	}
	//
	//	// 去除多余的 tag
	//	if len(tagResults) > cfg.Size {
	//		tagResults = tagResults[:cfg.Size]
	//	}
	//
	//	return json.Marshal(tagResults)
	//default:
	//	klog.Errorf("不支持的远端仓库类型“ %v", cfg.ImageFrom)
	//}
}

func (s *AgentController) SearchDockerhubTagInfo(ctx context.Context, req types.RemoteTagInfoSearchRequest) ([]byte, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s/", req.Namespace, req.Repository, req.Tag)
	resp, err := DoHttpRequest(url)
	if err != nil {
		return nil, err
	}

	var searchDockerhubImageResult types.SearchDockerhubTagInfoResult
	if err = json.Unmarshal(resp, &searchDockerhubImageResult); err != nil {
		return nil, err
	}
	return json.Marshal(types.CommonSearchTagInfoResult{
		Name:     req.Tag,
		FullSize: searchDockerhubImageResult.FullSize,
		Digest:   searchDockerhubImageResult.Digest,
		Images:   searchDockerhubImageResult.Images,
	})
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

func (s *AgentController) SearchTagInfo(ctx context.Context, req types.RemoteTagInfoSearchRequest) ([]byte, error) {
	switch req.Hub {
	case types.ImageHubDocker:
		return s.SearchDockerhubTagInfo(ctx, req)
	case types.ImageHubGCR:
		return s.SearchGCRagInfo(ctx, req)
	}
	return nil, nil
}

func (s *AgentController) SyncKubernetesTags(ctx context.Context, req types.KubernetesTagRequest) ([]byte, error) {
	if !req.SyncAll {
		url := fmt.Sprintf("https://api.github.com/repos/kubernetes/kubernetes/tags?per_page=10")
		return DoHttpRequest(url)
	}

	var allData []byte
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/repos/kubernetes/kubernetes/tags?per_page=100&page=%d", page)
		data, err := DoHttpRequest(url)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 || bytes.Equal(data, []byte("[]")) {
			break
		}
		allData = appendData(allData, data)

		page++
		time.Sleep(10 * time.Millisecond) // 避免请求过快
	}

	return allData, nil
}

func appendData(allData, newData []byte) []byte {
	// 情况1: 原始二进制直接拼接
	// return append(allData, newData...)

	// 情况2: JSON数组合并（更常见）
	if len(allData) == 0 {
		return newData
	}

	// 移除allData最后的']'和newData开头的'['
	if allData[len(allData)-1] == ']' && newData[0] == '[' {
		allData = allData[:len(allData)-1]
		newData = newData[1:]
		return append(append(allData, ','), newData...)
	}

	return append(allData, newData...)
}

func (s *AgentController) Run(ctx context.Context, workers int) error {
	// 注册 rainbow 代理
	if err := s.RegisterAgentIfNotExist(ctx); err != nil {
		return err
	}

	go s.startHeartbeat(ctx)
	go s.getNextWorkItems(ctx)
	go s.startSyncActionUsage(ctx)
	go s.startGC(ctx)

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 1*time.Second)
	}

	return nil
}

func (s *AgentController) startGC(ctx context.Context) {
	// 1小时尝试回收一次
	ticker := time.NewTicker(900 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.GarbageCollect(ctx); err != nil {
			klog.Errorf("GarbageCollect 失败: %v", err)
			continue
		}
		klog.Infof("GarbageCollect 完成")
	}
}

func (s *AgentController) GarbageCollect(ctx context.Context) error {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "plugin" {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			klog.Errorf("获取文件夹(%s)信息失败 %v, 忽略", entry.Name(), err)
			continue
		}

		// 回收指定时间的文件
		now := time.Now()
		if now.Sub(fileInfo.ModTime()) > 30*time.Minute {
			removeDir := filepath.Join(s.baseDir, fileInfo.Name())
			util.RemoveFile(removeDir)
			klog.Infof("任务文件 %s 已被回收", removeDir)
		} else {
			klog.Infof("任务文件 %s 还在有效期内，暂不回收", fileInfo.Name())
		}
	}

	return nil
}

func (s *AgentController) startSyncActionUsage(ctx context.Context) {
	rand.Seed(time.Now().UnixNano())

	// 15分钟同步一次
	ticker := time.NewTicker(900 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		agent, err := s.factory.Agent().GetByName(ctx, s.name)
		if err != nil {
			klog.Errorf("获取 agent 失败 %v 等待下次同步", err)
			continue
		}
		if len(agent.GithubUser) == 0 || len(agent.GithubRepository) == 0 || len(agent.GithubToken) == 0 {
			klog.Infof("agent(%s) 的 github 属性存在空值，忽略", agent.Name)
			continue
		}
		if agent.Status != model.RunAgentType {
			klog.Warningf("agent 处于未运行状态，忽略")
			continue
		}

		// TODO: 随机等待一段时间
		klog.Infof("开始同步 agent(%s) 的 usage", agent.Name)
		if err = s.syncActionUsage(ctx, *agent); err != nil {
			klog.Errorf("agent(%s) 同步 usage 失败 %v", agent.Name, err)
			continue
		}
		//klog.Infof("完成同步 agent(%s) 的 usage", agent.Name)
	}
}

func (s *AgentController) syncActionUsage(ctx context.Context, agent model.Agent) error {
	month := time.Now().Format("1")

	url := fmt.Sprintf("https://api.github.com/users/%s/settings/billing/usage?month=%s", agent.GithubUser, month)
	klog.Infof("当前 %s 月, 将通过请求 %s 获取本月账单", month, url)

	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest("", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", agent.GithubToken))
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error resp %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var ud UsageData
	if err = json.Unmarshal(data, &ud); err != nil {
		return err
	}

	var grossAmount float64 = 0
	for _, item := range ud.UsageItems {
		grossAmount += item.GrossAmount
	}

	rounded := math.Round(grossAmount*1000) / 1000
	klog.Infof("Agent(%s)当月截止目前已经使用 %d 美金", agent.Name, rounded)
	if agent.GrossAmount == rounded {
		klog.Infof("agent(%s) 的 grossAmount 未发生变化，等待下一次同步", agent.Name)
		return nil
	}

	return s.factory.Agent().UpdateByName(ctx, agent.Name, map[string]interface{}{"gross_amount": rounded})
}

type UsageData struct {
	UsageItems []UsageItem `json:"usageItems"`
}

type UsageItem struct {
	Date           time.Time `json:"date"`
	Product        string    `json:"product"`
	SKU            string    `json:"sku"`
	Quantity       float64   `json:"quantity"`
	UnitType       string    `json:"unitType"`
	PricePerUnit   float64   `json:"pricePerUnit"`
	GrossAmount    float64   `json:"grossAmount"`
	DiscountAmount float64   `json:"discountAmount"`
	NetAmount      float64   `json:"netAmount"`
	RepositoryName string    `json:"repositoryName"`
}

func (s *AgentController) startHeartbeat(ctx context.Context) {
	klog.Infof("启动 agent 心跳检测")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		old, err := s.factory.Agent().GetByName(ctx, s.name)
		if err != nil {
			klog.Error("failed to get agent status %v", err)
			continue
		}

		updates := map[string]interface{}{"last_transition_time": time.Now()}
		if old.Status != model.UnRunAgentType {
			if old.Status == model.UnknownAgentType {
				updates["status"] = model.RunAgentType
				updates["message"] = "Agent started posting status"
			}
		}

		if err = s.factory.Agent().UpdateByName(ctx, s.name, updates); err != nil {
			klog.Error("同步 agent(%s) 心跳失败%v", s.name, err)
		} else {
			klog.V(2).Infof("同步 agent(%s) 心跳成功 %v", s.name, updates)
		}
	}
}

func (s *AgentController) getNextWorkItems(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取未处理
		tasks, err := s.factory.Task().ListWithAgent(ctx, s.name, 0)
		if err != nil {
			klog.Error("failed to list tasks %v", err)
			continue
		}
		if len(tasks) == 0 {
			continue
		}

		for _, task := range tasks {
			s.queue.Add(fmt.Sprintf("%d/%d", task.Id, task.ResourceVersion))
		}
	}
}

func (s *AgentController) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *AgentController) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	klog.Infof("任务(%v)被调度到本节点，即将开始处理", key)
	taskId, resourceVersion, err := KeyFunc(key)
	if err != nil {
		s.handleErr(ctx, err, key)
	} else {
		_ = s.factory.Task().UpdateDirectly(ctx, taskId, map[string]interface{}{"status": "镜像初始化", "message": "初始化环境中", "process": 1})
		if err = s.factory.Task().CreateTaskMessage(ctx, &model.TaskMessage{TaskId: taskId, Message: "节点调度完成"}); err != nil {
			klog.Errorf("记录节点调度失败 %v", err)
		}
		if err = s.sync(ctx, taskId, resourceVersion); err != nil {
			if msgErr := s.factory.Task().CreateTaskMessage(ctx, &model.TaskMessage{TaskId: taskId, Message: fmt.Sprintf("同步失败，原因: %v", err)}); msgErr != nil {
				klog.Errorf("记录同步失败 %v", msgErr)
			}
			s.handleErr(ctx, err, key)
		}
	}
	return true
}

func (s *AgentController) GetOneAdminRegistry(ctx context.Context) (*model.Registry, error) {
	regs, err := s.factory.Registry().GetAdminRegistries(ctx)
	if err != nil {
		klog.Errorf("获取默认镜像仓库失败: %v", err)
		return nil, err
	}
	if len(regs) == 0 {
		klog.Errorf("no admin or default registry found")
		return nil, fmt.Errorf("no admin or default registry found")
	}

	// 随机分，暂时不考虑负载情况，后续优化
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(len(regs))
	t := regs[x]
	return &t, err
}

func (s *AgentController) makePluginConfig(ctx context.Context, task model.Task) (*rainbowconfig.PluginTemplateConfig, error) {
	taskId := task.Id

	var (
		registry *model.Registry
		err      error
	)
	// 未指定自定义参考时，使用默认仓库
	if task.RegisterId == 0 {
		registry, err = s.GetOneAdminRegistry(ctx)
	} else {
		registry, err = s.factory.Registry().Get(ctx, task.RegisterId)
	}
	if err != nil {
		klog.Error("failed to get registry %v", err)
		return nil, fmt.Errorf("failed to get registry %v", err)
	}

	pluginTemplateConfig := &rainbowconfig.PluginTemplateConfig{
		Default: rainbowconfig.DefaultOption{
			Time: time.Now().Unix(), // 注入时间戳，确保每次内容都不相同
		},
		Plugin: rainbowconfig.PluginOption{
			Callback:   s.callback,
			TaskId:     taskId,
			RegistryId: registry.Id,
			Synced:     true,
			Driver:     task.Driver,
			Arch:       task.Architecture,
		},
		Registry: rainbowconfig.Registry{
			Repository: registry.Repository,
			Namespace:  registry.Namespace,
			Username:   registry.Username,
			Password:   registry.Password,
		},
	}

	// 根据type判断是镜像列表推送还是k8s镜像组推送
	switch task.Type {
	case 0:
		tags, err := s.factory.Image().ListTags(ctx, db.WithTaskLike(taskId), db.WithErrorTask(task.OnlyPushError))
		if err != nil {
			klog.Errorf("获取任务所属 tags 失败 %v", err)
			return nil, err
		}

		var imageIds []int64
		imageMap := make(map[int64][]model.Tag)
		for _, tag := range tags {
			imageIds = append(imageIds, tag.ImageId)
			old, ok := imageMap[tag.ImageId]
			if ok {
				imageMap[tag.ImageId] = append(old, tag)
			} else {
				imageMap[tag.ImageId] = []model.Tag{tag}
			}
		}
		images, err := s.factory.Image().List(ctx, db.WithIDIn(imageIds...))
		if err != nil {
			klog.Errorf("获取任务所属镜像失败 %v", err)
			return nil, err
		}

		var img []rainbowconfig.Image
		for _, i := range images {
			ts, ok := imageMap[i.Id]
			if !ok {
				klog.Warningf("未能找到镜像(%s)的tags", i.Name)
				continue
			}
			var tagStr []string
			for _, tt := range ts {
				tagStr = append(tagStr, tt.Name)
			}
			img = append(img, rainbowconfig.Image{
				Name: i.Name,
				Id:   i.Id,
				Path: i.Path,
				Tags: tagStr,
			})
		}
		pluginTemplateConfig.Default.PushImages = true
		pluginTemplateConfig.Images = img
	case 1:
		pluginTemplateConfig.Default.PushKubernetes = true
		pluginTemplateConfig.Kubernetes.Version = task.KubernetesVersion
	}

	return pluginTemplateConfig, err
}

func (s *AgentController) sync(ctx context.Context, taskId int64, resourceVersion int64) error {
	task, err := s.factory.Task().GetOne(ctx, taskId, resourceVersion)
	if err != nil {
		if errors.IsNotUpdated(err) {
			return nil
		}
		return fmt.Errorf("failted to get one task %d %v", taskId, err)
	}
	klog.Infof("开始处理任务(%s),任务ID(%d)", task.Name, taskId)

	tplCfg, err := s.makePluginConfig(ctx, *task)
	cfg, err := yaml.Marshal(tplCfg)
	if err != nil {
		return err
	}

	taskIdStr := fmt.Sprintf("%d", taskId)

	destDir := filepath.Join(s.baseDir, taskIdStr)
	if err = util.EnsureDirectoryExists(destDir); err != nil {
		return err
	}
	if !util.IsDirectoryExists(destDir + "/plugin") {
		if err = util.Copy(s.baseDir+"/plugin", destDir); err != nil {
			return err
		}
	}

	git := util.NewGit(destDir+"/plugin", taskIdStr, taskIdStr+"-"+time.Now().String())
	if err = git.Checkout(); err != nil {
		return err
	}
	if err = util.WriteIntoFile(string(cfg), destDir+"/plugin/config.yaml"); err != nil {
		return err
	}
	if err = git.Push(); err != nil {
		return err
	}
	return nil
}

// TODO
func (s *AgentController) handleErr(ctx context.Context, err error, key interface{}) {
	if err == nil {
		return
	}
	klog.Error(err)
}

func (s *AgentController) RegisterAgentIfNotExist(ctx context.Context) error {
	if len(s.name) == 0 {
		return fmt.Errorf("agent name missing")
	}

	var err error
	_, err = s.factory.Agent().GetByName(ctx, s.name)
	if err == nil {
		return nil
	}
	_, err = s.factory.Agent().Create(ctx, &model.Agent{Name: s.name, Status: model.RunAgentType, Type: model.PublicAgentType, Message: "Agent started posting status"})
	return err
}

func KeyFunc(key interface{}) (int64, int64, error) {
	str, ok := key.(string)
	if !ok {
		return 0, 0, fmt.Errorf("failed to convert %v to string", key)
	}
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("parts length not 2")
	}

	taskId, err := strutil.ParseInt64(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to Parse taskId to Int64 %v", err)
	}
	resourceVersion, err := strutil.ParseInt64(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to Parse resourceVersion to Int64 %v", err)
	}

	return taskId, resourceVersion, nil
}
