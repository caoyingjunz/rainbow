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

type SearchRepositoriesResult struct {
	Code    int                                  `json:"code"`
	Result  []types.CommonSearchRepositoryResult `json:"result,omitempty"`
	Message string                               `json:"message,omitempty"`
}

type SearchOptions struct {
	baseURL string
	cfg     *config.Config

	accessKey string
	signature string

	Limit int
}

func NewSearchCommand() *cobra.Command {
	o := &SearchOptions{
		baseURL: baseURL,
		Limit:   10,
	}

	cmd := &cobra.Command{
		Use:   "search <image>",
		Short: "Search Docker Hub for images",
		Long:  `Search Docker Hub for images through the PixiuHub API.`,
		Example: `  pixiuctl search nginx
  pixiuctl search redis --limit 15`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(strings.Join(args, " ")))
		},
	}

	cmd.Flags().IntVar(&o.Limit, "limit", 10, "Maximum number of results (1-100)")

	return cmd
}

func (o *SearchOptions) Complete(cmd *cobra.Command, args []string) error {
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

func (o *SearchOptions) Validate() error {
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

func (o *SearchOptions) Run(query string) error {
	o.accessKey = o.cfg.Auth.AccessKey
	o.signature = signatureutil.GenerateSignature(
		map[string]string{
			"action":    "pullOrCacheRepo",
			"accessKey": o.accessKey,
		},
		[]byte(o.cfg.Auth.SecretKey))

	q := url.Values{}
	q.Set("query", query)
	q.Set("page_size", strconv.Itoa(o.Limit))

	apiURL := fmt.Sprintf("%s/api/v2/search/repositories?%s", o.baseURL, q.Encode())

	var result SearchRepositoriesResult
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
		return fmt.Errorf("search failed with code %d", result.Code)
	}

	printSearchResults(result.Result)
	return nil
}

func printSearchResults(items []types.CommonSearchRepositoryResult) {
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "No repositories found.")
		return
	}
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tSTARS\tOFFICIAL\tDESCRIPTION")
	for _, r := range items {
		desc := ""
		if r.ShortDesc != nil {
			desc = *r.ShortDesc
		}
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		official := "no"
		if r.IsOfficial {
			official = "yes"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", r.Name, r.Stars, official, desc)
	}
}
