package structurizationmodels

type StructurizedMessage struct {
	Explanation string `json:"explanation"`
}

type StructRequest struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
}
