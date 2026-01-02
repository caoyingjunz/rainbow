package rainbow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/member"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/google/uuid"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

// EnableChartRepo
// 1. 初始化项目
// 2. 创建用户
// 3. 关联用户到项目
func (s *ServerController) EnableChartRepo(ctx context.Context, req *types.EnableChartRepoRequest) error {
	if len(req.ProjectName) == 0 {
		req.ProjectName = req.UserName
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
		klog.Errorf("创建 harbor 项目失败 %v", err)
		return err
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
		klog.Errorf("创建 harbor 用户失败 %v", err)
		return err
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
		klog.Errorf("关联用户到项目失败 %v", err)
		return err
	}
	return nil
}

func (s *ServerController) ListCharts(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts", repoCfg.URL, repoCfg.Namespace, listOption.Project)

	httpClient := util.NewHttpClient(5*time.Second, url)
	httpClient.WithAuth(repoCfg.Username, repoCfg.Password)

	var cs []types.ChartInfo
	if err := httpClient.Get(url, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (s *ServerController) DeleteChart(ctx context.Context, chartReq types.ChartMetaRequest) error {
	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart)

	httpClient := util.NewHttpClient(5*time.Second, url)
	httpClient.WithAuth(repoCfg.Username, repoCfg.Password)

	if err := httpClient.Delete(url, nil); err != nil {
		return err
	}
	return nil
}

func (s *ServerController) ListChartTags(ctx context.Context, chartReq types.ChartMetaRequest, listOption types.ListOptions) (interface{}, error) {
	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart)

	httpClient := util.NewHttpClient(5*time.Second, url)
	httpClient.WithAuth(repoCfg.Username, repoCfg.Password)

	var cs []types.ChartVersion
	if err := httpClient.Get(url, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (s *ServerController) GetChartTag(ctx context.Context, chartReq types.ChartMetaRequest) (interface{}, error) {
	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts/%s/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart, chartReq.Version)

	httpClient := util.NewHttpClient(5*time.Second, url)
	httpClient.WithAuth(repoCfg.Username, repoCfg.Password)
	if err := httpClient.Get(url, nil); err != nil {
		return nil, err
	}

	var cs types.ChartDetail
	if err := httpClient.Get(url, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (s *ServerController) DeleteChartTag(ctx context.Context, chartReq types.ChartMetaRequest) error {
	repoCfg := s.cfg.Server.Harbor
	url := fmt.Sprintf("%s/api/%s/%s/charts/%s/%s", repoCfg.URL, repoCfg.Namespace, chartReq.Project, chartReq.Chart, chartReq.Version)

	httpClient := util.NewHttpClient(5*time.Second, url)
	httpClient.WithAuth(repoCfg.Username, repoCfg.Password)

	if err := httpClient.Delete(url, nil); err != nil {
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
	tmpFile := filepath.Join("/tmp", fmt.Sprintf("%s_%s_%s_%s", chartReq.Project, time.Now().Format("20060102_150405"), uuid.New().String()[:8], name))
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

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.SetBasicAuth(repoCfg.Username, repoCfg.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "Go-Harbor-Client/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	return nil
}
