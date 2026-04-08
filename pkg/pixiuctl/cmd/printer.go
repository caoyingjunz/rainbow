package cmd

import (
	"fmt"
	"os"
	"strings"
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

func PrintImagesTable(images []model.Image, output string) {
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	wide := strings.EqualFold(strings.TrimSpace(output), "wide")
	if wide {
		fmt.Fprintln(w, "NAME\tID\tIS_PUBLIC\tTAGS\tPULLS\tPROTECTED\tCREATED\tREPO")
		for _, i := range images {
			created := i.GmtCreate.Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%d\t%s\t%s\t%s\n", i.Name, i.Id, imagePublic(i.IsPublic), i.TagsCount, i.Pull, imageProtected(i.IsLocked), created, i.Mirror)
		}
		return
	}

	fmt.Fprintln(w, "NAME\tID\tIS_PUBLIC\tTAGS\tPULLS\tCREATED")
	for _, i := range images {
		created := i.GmtCreate.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%d\t%s\n", i.Name, i.Id, imagePublic(i.IsPublic), i.TagsCount, i.Pull, created)
	}
}

func imageProtected(locked bool) string {
	if locked {
		return "yes"
	}
	return "no"
}

func imagePublic(public bool) string {
	if public {
		return "yes"
	}
	return "no"
}

func PrintImageTagsTable(tags []model.Tag) {
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	if len(tags) == 0 {
		fmt.Fprintln(os.Stdout, "No tags found.")
		return
	}

	fmt.Fprintln(w, "TAG\tOS/ARCH\tSTATUS\tSIZE\tMODIFIED\tSOURCE\tPULL")
	for _, t := range tags {
		modified := t.GmtModified.Format("2006-01-02 15:04:05")
		source := t.Path
		pullCmd := "docker pull " + t.Mirror + ":" + t.Name
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", t.Name, t.Architecture, t.Status, t.ReadSize, modified, source, pullCmd)
	}
}
