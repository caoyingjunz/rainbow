package types

type IdMeta struct {
	ID int64 `uri:"Id" binding:"required"`
}

type TaskMeta struct {
	TaskId int64 `form:"task_id"`
}

type IdNameMeta struct {
	ID   int64  `uri:"Id" binding:"required" form:"id"`
	Name string `uri:"name" binding:"required" form:"name"`
}
