package chat_models

type VoiceRecognitionRequest struct {
	Bucket   string `json:"bucket_name"`
	Filename string `json:"file_id"`
}

type VoiceRecognitionResponse struct {
	Text string `json:"text"`
}
