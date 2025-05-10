package chat_models

type VoiceRecognitionRequest struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
}

type VoiceRecognitionResponse struct {
	Text string `json:"text"`
}
