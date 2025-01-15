package client

func (e *NotFoundError) Error() string {
	return e.Message
}

func (c *HetznerRobotClient) IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
