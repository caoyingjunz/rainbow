package types

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type IdMeta struct {
	ID int64 `uri:"Id" binding:"required"`
}

type NameMeta struct {
	Namespace string `uri:"namespace" binding:"required" form:"name"`
	Name      string `uri:"name" binding:"required" form:"name"`
}

type TaskMeta struct {
	TaskId int64 `form:"task_id"`
}

type UserMeta struct {
	UserId string `form:"user_id"`
}

type IdNameMeta struct {
	ID   int64  `uri:"Id" binding:"required" form:"id"`
	Name string `uri:"name" binding:"required" form:"name"`
}

type DownflowMeta struct {
	ImageId   int64  `form:"image_id"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
}

type Response struct {
	Code    int           `json:"code"`              // 返回的状态码
	Result  []model.Image `json:"result,omitempty"`  // 正常返回时的数据，可以为任意数据结构
	Message string        `json:"message,omitempty"` // 异常返回时的错误信息
}

const (
	SyncImageInitializing = "Initializing"
	SyncImageRunning      = "Running"
	SyncImageError        = "Error"
	SyncImageComplete     = "Completed"
)

const (
	SkopeoDriver = "skopeo"
	DockerDriver = "docker"
)

const (
	SyncTaskInitializing = "initializing"
)

const (
	ImageHubDocker = "dockerhub"
	ImageHubGCR    = "gcr"
	ImageHubQuay   = "quay.io"
)

type SearchResult struct {
	Result     []byte
	ErrMessage string
	StatusCode int
}

type ImageTag struct {
	Features     string    `json:"features"`
	Variant      *string   `json:"variant"` // 可能是 null
	Digest       string    `json:"digest"`
	OS           string    `json:"os"`
	OSFeatures   string    `json:"os_features"`
	OSVersion    *string   `json:"os_version"` // 可能是 null
	Size         int64     `json:"size"`
	Status       string    `json:"status"`
	LastPulled   time.Time `json:"last_pulled"`
	LastPushed   time.Time `json:"last_pushed"`
	Architecture string    `json:"architecture"`
}

type HubSearchResponse struct {
	Count    int                `json:"count"`
	Next     string             `json:"next"`
	Previous string             `json:"previous"`
	Results  []RepositoryResult `json:"results"`
}

type RepositoryResult struct {
	RepoName         string `json:"repo_name"`
	ShortDescription string `json:"short_description"`
	StarCount        int    `json:"star_count"`
	PullCount        int64  `json:"pull_count"` // 使用 int64 因为拉取计数可能非常大
	RepoOwner        string `json:"repo_owner"`
	IsAutomated      bool   `json:"is_automated"`
	IsOfficial       bool   `json:"is_official"`
}

type HubTagResponse struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"` // 可能是 null 或字符串
	Results  []TagResult `json:"results"`
}

type TagResult struct {
	Images              []Image   `json:"images,omitempty"`
	LastUpdated         time.Time `json:"last_updated,omitempty"`
	LastUpdater         int64     `json:"last_updater,omitempty"`
	LastUpdaterUsername string    `json:"last_updater_username,omitempty"`
	Name                string    `json:"name,omitempty"`
	Repository          int64     `json:"repository,omitempty"`
	FullSize            int64     `json:"full_size,omitempty"`
	V2                  bool      `json:"v2,omitempty"`
	TagStatus           string    `json:"tag_status,omitempty"`
	TagLastPulled       time.Time `json:"tag_last_pulled,omitempty"`
	TagLastPushed       time.Time `json:"tag_last_pushed,omitempty"`
	MediaType           string    `json:"media_type,omitempty"`
	ContentType         string    `json:"content_type,omitempty"`
	Digest              string    `json:"digest,omitempty"`
}

type Image struct {
	Features     string    `json:"features,omitempty"`
	Variant      *string   `json:"variant,omitempty"` // 可能是 null
	Digest       string    `json:"digest,omitempty"`
	OS           string    `json:"os,omitempty"`
	OSFeatures   string    `json:"os_features,omitempty"`
	OSVersion    *string   `json:"os_version,omitempty"` // 可能是 null
	Size         int64     `json:"size,omitempty"`
	Status       string    `json:"status,omitempty"`
	LastPulled   time.Time `json:"last_pulled,omitempty"`
	LastPushed   time.Time `json:"last_pushed,omitempty"`
	Architecture string    `json:"architecture,omitempty"`
}
