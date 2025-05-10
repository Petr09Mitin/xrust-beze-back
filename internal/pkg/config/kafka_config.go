package config

type Kafka struct {
	Version                             string   `mapstructure:"version"`
	Addresses                           []string `mapstructure:"addresses"`
	StudyMaterialTopic                  string   `mapstructure:"study_material_topic,omitempty"`
	VoiceRecognitionNewVoiceTopic       string   `mapstructure:"voice_recognition_new_voice_topic,omitempty"`
	VoiceRecognitionVoiceProcessedTopic string   `mapstructure:"voice_recognition_voice_processed_topic,omitempty"`
}
