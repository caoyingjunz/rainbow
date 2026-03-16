package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

func PrintTable(registries []model.Registry) {
	// 使用 tabwriter 对齐输出
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tID\tCREATED")
	for _, r := range registries {
		created := r.GmtCreate.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%d\t%s\n", r.Name, r.Id, created)
	}
}

func PrintImagesTable(images []model.Image) {
	// 使用 tabwriter 对齐输出
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tID\tTAGS\tCREATED")
	for _, i := range images {
		created := i.GmtCreate.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", i.Name, i.Id, i.TagsCount, created)
	}
}
