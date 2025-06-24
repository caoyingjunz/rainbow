package types

import "github.com/caoyingjunz/rainbow/pkg/db/model"

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
	SyncTaskInitializing = "initializing"
)

type SearchResult struct {
	Result     []byte
	ErrMessage string
	StatusCode int
}
