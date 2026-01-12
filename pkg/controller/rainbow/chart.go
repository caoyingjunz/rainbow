package rainbow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/member"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/google/uuid"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

func (s *ServerController) preEnableChartRepo(ctx context.Context, req *types.EnableChartRepoRequest) error {
	_, err := s.factory.Task().GetUserBy(ctx, db.WithUser(req.UserId))
	if err == nil {
		return nil
	}
	klog.Errorf("获取用户失败 %v", err)
	return fmt.Errorf("获取用户属性失败，请联系管理员同步用户信息")
}

// EnableChartRepo
// 1. 初始化项目
// 2. 创建用户
// 3. 关联用户到项目
func (s *ServerController) EnableChartRepo(ctx context.Context, req *types.EnableChartRepoRequest) error {
	if err := s.preEnableChartRepo(ctx, req); err != nil {
		return err
	}

	if len(req.ProjectName) == 0 {
		req.ProjectName = strings.ToLower(req.UserName)
	}
	// 创建关联项目
	if _, err := s.chartRepoAPI.Project.CreateProject(ctx, &project.CreateProjectParams{
		Project: &models.ProjectReq{
			Metadata: &models.ProjectMetadata{
				Public: fmt.Sprintf("%t", req.Public),
			},
			ProjectName: req.ProjectName,
		},
	}); err != nil {
		if !strings.Contains(err.Error(), "createProjectConflict") {
			klog.Errorf("创建 harbor 项目失败 %v", err)
			return err
		}
		klog.Warningf("项目 %s 已经存在", req.ProjectName)
	}

	// 创建用户
	if _, err := s.chartRepoAPI.User.CreateUser(ctx, &user.CreateUserParams{
		UserReq: &models.UserCreationReq{
			Username: req.UserName,
			Password: req.Password,
			Comment:  "PixiuHub",
			Email:    req.Email,
			Realname: req.UserName,
		},
	}); err != nil {
		if !strings.Contains(err.Error(), "createUserConflict") {
			klog.Errorf("创建 harbor 用户失败 %v", err)
			return err
		}
		klog.Warningf("用户 %s 已经存在", req.UserName)
	}

	// 关联用户到项目
	if _, err := s.chartRepoAPI.Member.CreateProjectMember(ctx, &member.CreateProjectMemberParams{
		ProjectNameOrID: req.ProjectName,
		ProjectMember: &models.ProjectMember{
			RoleID: 4,
			MemberUser: &models.UserEntity{
				Username: req.UserName,
			},
		},
	}); err != nil {
		if !strings.Contains(err.Error(), "createProjectMemberConflict") {
			klog.Errorf("创建 harbor 用户失败 %v", err)
			return err
		}
		klog.Warningf("用户关系关联 %s 已经存在", req.UserName)
	}

	// 修改为启用状态
	if err := s.factory.Task().UpdateUserBy(ctx, map[string]interface{}{"enable_chart": true}, db.WithUser(req.UserId)); err != nil {
		klog.Errorf("更新用户启用状态失败 %v", err)
		return err
	}
	return nil
}

func (s *ServerController) GetChartStatus(ctx context.Context, req *types.ChartMetaRequest) (interface{}, error) {
	// 项目名称和用户名相同
	userName := req.Project
	ojb, err := s.factory.Task().GetUserBy(ctx, db.WithName(userName))
	if err != nil {
		return nil, err
	}

	return struct {
		EnableChart bool `json:"enable_chart"`
	}{EnableChart: ojb.EnableChart}, nil
}

func (s *ServerController) ListCharts(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	repoCfg := s.cfg.Server.Harbor

	var cs []types.ChartInfo
	httpClient := util.HttpClientV2{
		URL: fmt.Sprintf("%s/api/%s/%s/charts", repoCfg.URL, repoCfg.Namespace, listOption.Project),
	}
	err := httpClient.Method("GET").
		WithTimeout(5*time.Second).
		WithAuth(repoCfg.Username, repoCfg.Password).
		Do(&cs)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (s *ServerController) DeleteChart(ctx context.Context, chartReq types.ChartMetaRequest) error {
	repoCfg := s.cfg.Server.Harbor

	httpClient := util.HttpClientV2{
		URL: fmt.Sprintf("%s/api/%s/%s/charts/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart),
	}
	err := httpClient.Method("DELETE").
		WithTimeout(5*time.Second).
		WithAuth(repoCfg.Username, repoCfg.Password).
		Do(nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerController) ListChartTags(ctx context.Context, chartReq types.ChartMetaRequest, listOption types.ListOptions) (interface{}, error) {
	repoCfg := s.cfg.Server.Harbor

	url := fmt.Sprintf("%s/api/%s/%s/charts/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart)
	var cs []types.ChartVersion
	httpClient := util.HttpClientV2{
		URL: url,
	}
	err := httpClient.Method("GET").
		WithTimeout(5*time.Second).
		WithAuth(repoCfg.Username, repoCfg.Password).
		Do(&cs)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (s *ServerController) GetChartTag(ctx context.Context, chartReq types.ChartMetaRequest) (interface{}, error) {
	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts/%s/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart, chartReq.Version)

	var cs types.ChartDetail
	httpClient := util.HttpClientV2{
		URL: url,
	}
	err := httpClient.Method("GET").
		WithTimeout(5*time.Second).
		WithAuth(repoCfg.Username, repoCfg.Password).
		Do(&cs)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (s *ServerController) DeleteChartTag(ctx context.Context, chartReq types.ChartMetaRequest) error {
	repoCfg := s.cfg.Server.Harbor

	url := fmt.Sprintf("%s/api/%s/%s/charts/%s/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart, chartReq.Version)
	httpClient := util.HttpClientV2{
		URL: url,
	}
	err := httpClient.Method("DELETE").
		WithTimeout(5*time.Second).
		WithAuth(repoCfg.Username, repoCfg.Password).
		Do(nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerController) UploadChart(ctx *gin.Context, chartReq types.ChartMetaRequest) error {
	f, err := ctx.FormFile("chart")
	if err != nil {
		return err
	}

	name := f.Filename
	tmpFile := filepath.Join("/tmp", fmt.Sprintf("upload_%s_%s_%s_%s", chartReq.Project, time.Now().Format("20060102_150405"), uuid.New().String()[:8], name))
	if err = ctx.SaveUploadedFile(f, tmpFile); err != nil {
		return err
	}

	// 清理临时文件
	defer func() {
		if err = os.RemoveAll(tmpFile); err != nil {
			fmt.Printf("清理临时文件失败: %v\n", err)
		}
	}()

	return s.uploadChart(chartReq.Project, tmpFile)
}

func (s *ServerController) uploadChart(project string, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filename := filepath.Base(filePath)
	part, err := writer.CreateFormFile("chart", filename)
	if err != nil {
		return fmt.Errorf("创建表单文件字段失败: %w", err)
	}

	copied, err := io.Copy(part, f)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}
	klog.Infof("已读取文件大小: %d 字节\n", copied)
	if err = writer.Close(); err != nil {
		return fmt.Errorf("关闭表单写入器失败: %w", err)
	}

	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts", repoCfg.URL, repoCfg.Namespace, project)
	httpClient := util.HttpClientV2{
		URL: url,
	}
	var cs types.ChartSaved
	err = httpClient.Method("POST").
		WithTimeout(30*time.Second).
		WithHeader(map[string]string{"Accept": "application/json", "Content-Type": writer.FormDataContentType(), "User-Agent": "Go-Harbor-Client/1.0"}).
		WithBody(body).
		WithAuth(repoCfg.Username, repoCfg.Password).
		Do(&cs)
	if err != nil {
		return err
	}

	// 判断是否已保存
	if cs.Saved {
		return nil
	}
	return fmt.Errorf("chart not saved")
}

func (s *ServerController) DownloadChart(ctx *gin.Context, chartReq types.ChartMetaRequest) (string, string, error) {
	repoCfg := s.cfg.Server.Harbor

	chartName := fmt.Sprintf("%s-%s.tgz", chartReq.Chart, chartReq.Version)
	url := fmt.Sprintf("%s/%s/%s/charts/%s", repoCfg.URL, "charts", chartReq.Project, chartName)

	filename := filepath.Join("/tmp", fmt.Sprintf("donwload_%s_%s_%s_%s", chartReq.Project, time.Now().Format("20060102_150405"), uuid.New().String()[:8], chartName))
	httpClient := util.HttpClientV2{
		URL: url,
	}
	err := httpClient.Method("GET").
		WithTimeout(30*time.Second).
		WithAuth(repoCfg.Username, repoCfg.Password).
		WithFile(filename).
		Do(nil)
	if err != nil {
		return "", "", err
	}

	return chartName, filename, nil
}
