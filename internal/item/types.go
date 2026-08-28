package item

type ItemType string
type TaskStatus string
type ReminderStatus string

const (
	TypeNote     ItemType = "note"
	TypeTask     ItemType = "task"
	TypeReminder ItemType = "reminder"

	TaskOpenStatus     TaskStatus = "open"
	TaskDoneStatus     TaskStatus = "done"
	TaskArchivedStatus TaskStatus = "archived"

	ReminderPendingStatus   ReminderStatus = "pending"
	ReminderFiredStatus     ReminderStatus = "fired"
	ReminderDismissedStatus ReminderStatus = "dismissed"
)

func IsValidItemType(t string) bool {
	switch ItemType(t) {
	case TypeNote, TypeTask, TypeReminder:
		return true
	default:
		return false
	}
}

func IsValidTaskStatus(s string) bool {
	switch TaskStatus(s) {
	case TaskOpenStatus, TaskDoneStatus, TaskArchivedStatus:
		return true
	default:
		return false
	}
}

func IsValidReminderStatus(s string) bool {
	switch ReminderStatus(s) {
	case ReminderPendingStatus, ReminderFiredStatus, ReminderDismissedStatus:
		return true
	default:
		return false
	}
}
