package tgCommands

type TgCommand string

const (
	Start          string = "/start"
	NewTask        string = "/newTask"
	DoneTask       string = "/doneTask"
	CleanDoneTasks string = "/cleanDoneTasks"
	CleanAllTasks  string = "/cleanAllTasks"
	Done           string = "/done"
)
