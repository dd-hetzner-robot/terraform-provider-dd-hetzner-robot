package client

func (e *NotFoundError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) error {
	return &NotFoundError{Message: message}
}

func (c *HetznerRobotClient) IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
