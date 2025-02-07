package types

type IdMeta struct {
	ID int64 `uri:"Id" binding:"required"`
}

type TaskMeta struct {
	TaskId int64 `form:"task_id"`
}
