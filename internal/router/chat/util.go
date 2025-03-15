package chat_delivery

func DecodeToMessagePresenter(data []byte) (*MessagePresenter, error) {
	var msg MessagePresenter
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func DecodeToMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
