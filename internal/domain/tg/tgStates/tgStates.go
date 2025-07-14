package tgStates

type TgUserState int8

const (
	Undefined TgUserState = iota
	Default
	WaitAddTask
	WaitDoneTask
	AskNewTaskName
	AskDoneTaskName
	CleanDoneTasks
	CleanAllTasks
)
