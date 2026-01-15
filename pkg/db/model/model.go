package model

var models = make([]interface{}, 0)

func register(model ...interface{}) {
	models = append(models, model...)
}

func GetMigrationModels() []interface{} {
	return models
}

func GetAgentRunningStatus() []string {
	return []string{
		DeletingAgentType,
		StartingAgentType,
		StoppingAgentType,
		RestartingAgentType,
		UpgradeAgentType,
		OfflineAgentType,
	}
}
