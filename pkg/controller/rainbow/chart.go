package rainbow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
	"github.com/caoyingjunz/rainbow/pkg/util/tokenutil"
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

	// 创建项目
	if err := s.CreateRepoProject(ctx, req.ProjectName, req.Public); err != nil {
		klog.Errorf("创建项目(%s)失败 %v", req.ProjectName, err)
		return err
	}
	// 创建用户
	if err := s.CreateRepoUser(ctx, req); err != nil {
		klog.Errorf("创建用户(%s)失败 %v", req.UserName, err)
		return err
	}
	// 创建用户关联
	if err := s.CreateProjectMember(ctx, req); err != nil {
		klog.Errorf("关联项目用户(%s)失败 %v", req.UserName, err)
		return err
	}

	return nil
}

func (s *ServerController) EnableChartRepo2(ctx context.Context, req *types.EnableChartRepoRequest) error {
	if err := s.preEnableChartRepo(ctx, req); err != nil {
		return err
	}

	if len(req.ProjectName) == 0 {
		req.ProjectName = strings.ToLower(req.UserName)
	}
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

func (s *ServerController) DisableAndDeleteChartRepo(ctx context.Context, req *types.EnableChartRepoRequest) error {
	if len(req.ProjectName) == 0 {
		req.ProjectName = strings.ToLower(req.UserName)
	}

	if err := s.factory.Task().UpdateUserBy(ctx, map[string]interface{}{"enable_chart": false}, db.WithUser(req.UserId)); err != nil {
		klog.Errorf("更新用户启用状态失败 %v", err)
		return err
	}

	if _, err := s.chartRepoAPI.Member.DeleteProjectMember(ctx, &member.DeleteProjectMemberParams{ProjectNameOrID: req.ProjectName}); err != nil {
		klog.Warningf("接触项目(%s)关联失败 %v", req.ProjectName, err)
	}

	return nil
}

func (s *ServerController) GetChartStatus(ctx context.Context, req *types.ChartMetaRequest) (interface{}, error) {
	// 项目名称和用户名相同
	userName := req.Project
	ojb, err := s.factory.Task().GetUserBy(ctx, db.WithName(userName))
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Warningf("用户信息未同步，返回默认值")
			return struct {
				EnableChart bool `json:"enable_chart"`
			}{EnableChart: false}, nil
		}
		return nil, err
	}

	return struct {
		EnableChart bool `json:"enable_chart"`
	}{EnableChart: ojb.EnableChart}, nil
}

func (s *ServerController) CreateRepoProject(ctx context.Context, projectName string, public bool) error {
	repoCfg := s.cfg.Server.Harbor

	data, _ := json.Marshal(map[string]interface{}{
		"project_name": projectName,
		"public":       public,
	})
	httpClient := util.HttpClientV2{URL: fmt.Sprintf("%s/api/v2.0/projects", repoCfg.URL)}
	err := httpClient.Method("POST").
		WithTimeout(5*time.Second).
		WithHeader(map[string]string{"Content-Type": "application/json"}).
		WithAuth(repoCfg.Username, repoCfg.Password).
		WithBody(bytes.NewBuffer(data)).
		Do(nil)
	if err != nil {
		switch err.Code {
		case http.StatusConflict:
			klog.Infof("项目(%s)已存在", projectName)
			return nil
			//return fmt.Errorf("项目(%s)已存在", projectName)
		}
		return err
	}

	return nil
}

func (s *ServerController) CreateRepoUser(ctx context.Context, req *types.EnableChartRepoRequest) error {
	repoCfg := s.cfg.Server.Harbor

	data, _ := json.Marshal(map[string]interface{}{
		"username": req.UserName,
		"password": req.Password,
		"email":    req.Email,
		"realname": req.UserName,
		"comment":  "PixiuHub",
	})
	httpClient := util.HttpClientV2{URL: fmt.Sprintf("%s/api/v2.0/users", repoCfg.URL)}
	err := httpClient.Method("POST").
		WithTimeout(5*time.Second).
		WithHeader(map[string]string{"Content-Type": "application/json"}).
		WithAuth(repoCfg.Username, repoCfg.Password).
		WithBody(bytes.NewBuffer(data)).
		Do(nil)
	if err != nil {
		switch err.Code {
		case http.StatusConflict:
			klog.Infof("用户(%s)或邮箱(%s)已存在", req.UserName, req.Email)
			return nil
		}
		return err
	}

	return nil
}

func (s *ServerController) CreateProjectMember(ctx context.Context, req *types.EnableChartRepoRequest) error {
	repoCfg := s.cfg.Server.Harbor

	data, _ := json.Marshal(&models.ProjectMember{
		RoleID: 4,
		MemberUser: &models.UserEntity{
			Username: req.UserName,
		},
	})
	httpClient := util.HttpClientV2{URL: fmt.Sprintf("%s/api/v2.0/projects/%s/members", repoCfg.URL, req.ProjectName)}
	err := httpClient.Method("POST").
		WithTimeout(5*time.Second).
		WithHeader(map[string]string{"Content-Type": "application/json"}).
		WithAuth(repoCfg.Username, repoCfg.Password).
		WithBody(bytes.NewBuffer(data)).
		Do(nil)
	if err != nil {
		switch err.Code {
		case http.StatusConflict:
			klog.Infof("项目(%s)关联用户已存在", req.UserName)
			return nil
		}
		return err
	}

	return nil
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
	f, err := ctx.FormFile("file")
	if err != nil {
		klog.Errorf("获取chart文件失败 %v", err)
		return err
	}

	klog.Infof("已成功获取chart文件")
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
		klog.Infof("读取文件大小失败: %v", err)
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

	klog.Infof("chart %s 上传完成", filename)
	// 判断是否已保存
	if cs.Saved {
		return nil
	}
	return fmt.Errorf("chart not saved")
}

func (s *ServerController) DownloadChart(ctx *gin.Context, chartReq types.ChartMetaRequest) (string, string, error) {
	if err := s.ValidateToken(ctx); err != nil {
		klog.Errorf("验证token失败 %v", err)
		return "", "", err
	}

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

const JWTKey = "pixiuHub"

func (s *ServerController) GetToken(ctx context.Context, req *types.ChartMetaRequest) (interface{}, error) {
	token, err := tokenutil.GenerateToken("", req.Project, []byte(JWTKey))
	if err != nil {
		klog.Errorf("生成token失败 %v", err)
		return nil, err
	}
	return token, nil
}

func (s *ServerController) ValidateToken(ctx *gin.Context) error {
	token, err := s.extractToken(ctx)
	if err != nil {
		return err
	}
	_, err = tokenutil.ParseToken(token, []byte(JWTKey))
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerController) extractToken(c *gin.Context) (string, error) {
	emptyFunc := func(t string) bool { return len(t) == 0 }

	token := c.GetHeader("Authorization")
	if emptyFunc(token) {
		return "", fmt.Errorf("authorization header is not provided")
	}
	fields := strings.Fields(token)
	if len(fields) != 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}
	if fields[0] != "Bearer" {
		return "", fmt.Errorf("unsupported authorization type")
	}

	return fields[1], nil
}
