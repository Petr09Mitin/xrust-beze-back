package config

type Kafka struct {
	Version            string   `mapstructure:"version"`
	Addresses          []string `mapstructure:"addresses"`
	StudyMaterialTopic string   `mapstructure:"study_material_topic,omitempty"`
}
