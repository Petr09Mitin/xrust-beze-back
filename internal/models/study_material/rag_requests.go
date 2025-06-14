package study_material_models

type NewMaterialRAGData struct {
	Key     string `json:"key"`
	MongoID string `json:"mongo_id"`
}

type DeleteMaterialRAGData struct {
	Key     string `json:"key"`
	MongoID string `json:"mongo_id"`
}
