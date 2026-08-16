package item

import (
	"fmt"
	"strings"
	"time"
)

type Frontmatter struct {
	ID      string   `yaml:"id"`
	Type    string   `yaml:"type"`
	Created string   `yaml:"created"`
	Tags    []string `yaml:"tags"`
	Related []string `yaml:"related"`

	Due      string `yaml:"due"`
	Status   string `yaml:"status"`
	RemindAt string `yaml:"remind_at"`
}

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

func (f Frontmatter) Validate() error {
	if err := f.generalValidation(); err != nil {
		return err
	}
	switch ItemType(f.Type) {
	case TypeNote:
		return f.noteValidation()
	case TypeTask:
		return f.taskValidation()
	case TypeReminder:
		return f.reminderValidation()
	}
	return nil
}

func (f Frontmatter) generalValidation() error {
	if isEmpty(f.ID) {
		return fmt.Errorf("id can't be empty")
	}
	if !isValidItemType(f.Type) {
		return fmt.Errorf("invalid type: %q (should be: note, task or reminder)", f.Type)
	}
	if !isValidDate(time.RFC3339, f.Created) {
		return fmt.Errorf("invalid created date format: %q (should be: %s)", f.Created, time.RFC3339)
	}
	return nil
}

func (f Frontmatter) noteValidation() error {
	if !isEmpty(f.Due) {
		return fmt.Errorf("note must not have a due date")
	}
	if !isEmpty(f.Status) {
		return fmt.Errorf("note must not have a status")
	}
	if !isEmpty(f.RemindAt) {
		return fmt.Errorf("note must not have a remind_at")
	}
	return nil
}

func (f Frontmatter) taskValidation() error {
	if !isValidDate(time.DateOnly, f.Due) {
		return fmt.Errorf("invalid due date format: %q (should be: %s)", f.Due, time.DateOnly)
	}
	if !isValidTaskStatus(f.Status) {
		return fmt.Errorf("invalid task status: %q (should be: open, done, archived)", f.Status)
	}
	if !isEmpty(f.RemindAt) {
		return fmt.Errorf("task must not have a remind_at")
	}
	return nil
}

func (f Frontmatter) reminderValidation() error {
	if !isValidDate(time.RFC3339, f.RemindAt) {
		return fmt.Errorf("invalid remind_at date format: %q (should be: %s)", f.RemindAt, time.RFC3339)
	}
	if !isValidReminderStatus(f.Status) {
		return fmt.Errorf("invalid reminder status: %q (should be: pending, fired, dismissed)", f.Status)
	}
	if !isEmpty(f.Due) {
		return fmt.Errorf("reminder must not have a due date")
	}
	return nil
}

func isValidItemType(t string) bool {
	switch ItemType(t) {
	case TypeNote, TypeTask, TypeReminder:
		return true
	default:
		return false
	}
}

func isValidTaskStatus(s string) bool {
	switch TaskStatus(s) {
	case TaskOpenStatus, TaskDoneStatus, TaskArchivedStatus:
		return true
	default:
		return false
	}
}

func isValidReminderStatus(s string) bool {
	switch ReminderStatus(s) {
	case ReminderPendingStatus, ReminderFiredStatus, ReminderDismissedStatus:
		return true
	default:
		return false
	}
}

func isValidDate(layout string, d string) bool {
	_, err := time.Parse(layout, d)
	return err == nil
}

func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
