package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/signatureutil"
)

type SearchTagsResult struct {
	Code    int                         `json:"code"`
	Result  types.CommonSearchTagResult `json:"result,omitempty"`
	Message string                      `json:"message,omitempty"`
}

type LsOptions struct {
	baseURL string
	cfg     *config.Config

	accessKey string
	signature string

	Limit int
	Query string
}

func NewLsCommand() *cobra.Command {
	o := &LsOptions{
		baseURL: baseURL,
		Limit:   10,
	}

	cmd := &cobra.Command{
		Use:   "ls <image>",
		Short: "List tags of a Docker Hub image",
		Long:  `List repository tags from Docker Hub through the PixiuHub API.`,
		Example: `  pixiuctl ls nginx
  pixiuctl ls library/nginx --limit 20
  pixiuctl ls redis --query alpine`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			cmdutil.CheckErr(o.Complete(cmd))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(args[0]))
		},
	}

	cmd.Flags().IntVar(&o.Limit, "limit", 10, "Maximum number of tag results (1-100)")
	cmd.Flags().StringVar(&o.Query, "query", "", "Tag name selector, e.g. v1 or alpine")

	return cmd
}

func (o *LsOptions) Complete(cmd *cobra.Command) error {
	configFile, err := cmd.Root().PersistentFlags().GetString("configFile")
	if err != nil {
		return err
	}
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return err
	}
	o.cfg = cfg
	if o.cfg.Default != nil && len(o.cfg.Default.URL) != 0 {
		o.baseURL = o.cfg.Default.URL
	}
	return nil
}

func (o *LsOptions) Validate() error {
	if o.cfg.Auth == nil {
		return fmt.Errorf("配置文件缺少 Auth")
	}
	if len(o.cfg.Auth.AccessKey) == 0 {
		return fmt.Errorf("配置文件缺少 auth.access_key")
	}
	if len(o.cfg.Auth.SecretKey) == 0 {
		return fmt.Errorf("配置文件缺少 auth.secret_key")
	}
	if o.Limit < 1 {
		return fmt.Errorf("--limit must be at least 1")
	}
	if o.Limit > 100 {
		return fmt.Errorf("--limit must be at most 100")
	}
	return nil
}

func (o *LsOptions) Run(imageRepo string) error {
	o.accessKey = o.cfg.Auth.AccessKey
	o.signature = signatureutil.GenerateSignature(
		map[string]string{
			"action":    "pullOrCacheRepo",
			"accessKey": o.accessKey,
		},
		[]byte(o.cfg.Auth.SecretKey))

	namespace, repository := parseImageRepo(imageRepo)

	q := url.Values{}
	q.Set("namespace", namespace)
	q.Set("repository", repository)
	q.Set("page_size", strconv.Itoa(o.Limit))
	if strings.TrimSpace(o.Query) != "" {
		q.Set("query", strings.TrimSpace(o.Query))
	}

	apiURL := fmt.Sprintf("%s/api/v2/search/repositories/tags?%s", o.baseURL, q.Encode())

	var result SearchTagsResult
	httpClient := util.HttpClientV2{URL: apiURL}
	if err := httpClient.Method(http.MethodGet).
		WithTimeout(60 * time.Second).
		WithHeader(map[string]string{
			"X-ACCESS-KEY":  o.accessKey,
			"Authorization": o.signature,
		}).
		Do(&result); err != nil {
		return err
	}
	if result.Code != 200 {
		if result.Message != "" {
			return fmt.Errorf("%s", result.Message)
		}
		return fmt.Errorf("ls failed with code %d", result.Code)
	}

	printTagResults(result.Result)
	return nil
}

func parseImageRepo(imageRepo string) (string, string) {
	parts := strings.Split(strings.TrimSpace(imageRepo), "/")
	if len(parts) <= 1 {
		return "library", imageRepo
	}
	return parts[0], strings.Join(parts[1:], "/")
}

func printTagResults(result types.CommonSearchTagResult) {
	if len(result.TagResult) == 0 {
		fmt.Fprintln(os.Stdout, "No tags found.")
		return
	}

	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "TAG\tSIZE\tLAST_MODIFIED\tDIGEST")
	for _, t := range result.TagResult {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", t.Name, t.Size, t.LastModified, t.ManifestDigest)
	}
}
