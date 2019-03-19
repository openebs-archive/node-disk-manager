package udev

const (
	UdevEventAdd    = "add"
	UdevEventRemove = "remove"
	UdevEventChange = "change"
)

// TODO trigger a udev event, so that disk attach, detach can be simulated
func TriggerEvent(event, device string) {

}

// TODO Create a fake disk
func CreateDisk(name string) {

}
