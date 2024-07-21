package notify

type Notify interface {
	PushNotify(content, title string) error
}
