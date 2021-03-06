package scheduler

import (
	"fmt"
	"sort"
	"sync"

	"github.com/xitongsys/guery/config"
)

type TaskList []*Task

func (self TaskList) Len() int      { return len(self) }
func (self TaskList) Swap(i, j int) { self[i], self[j] = self[j], self[i] }
func (self TaskList) Less(i, j int) bool {
	if self[i].Runtime.Priority == self[j].Runtime.Priority {
		return self[i].CommitTime.After(self[j].CommitTime)
	}
	return self[i].Runtime.Priority > self[j].Runtime.Priority
}

type Queue struct {
	sync.Mutex
	Name         string
	MaxQueueSize int32
	Tasks        TaskList
}

func NewQueue(name string) *Queue {
	return &Queue{
		Name:         name,
		MaxQueueSize: config.Conf.Runtime.MaxQueueSize,
	}
}

func (self *Queue) Top() *Task {
	self.Lock()
	defer self.Unlock()

	ln := len(self.Tasks)
	if ln <= 0 {
		return nil
	}
	return self.Tasks[ln-1]
}

func (self *Queue) Pop() *Task {
	self.Lock()
	defer self.Unlock()

	ln := len(self.Tasks)
	if ln > 0 {
		task := self.Tasks[ln-1]
		self.Tasks = self.Tasks[:ln-1]
		return task
	}
	return nil
}

func (self *Queue) Delete(task *Task) error {
	self.Lock()
	defer self.Unlock()

	i, ln := 0, len(self.Tasks)
	for i = 0; i < ln && self.Tasks[i].TaskId != task.TaskId; i++ {
	}
	if i >= ln {
		return fmt.Errorf("task not in this list")
	}

	for j := i; j < ln-1; j++ {
		self.Tasks[i] = self.Tasks[i+1]
	}
	self.Tasks = self.Tasks[:ln-1]
	return nil
}

func (self *Queue) AddForce(task *Task) error {
	self.Lock()
	defer self.Unlock()

	ln := len(self.Tasks)
	if ln >= int(self.MaxQueueSize) {
		self.Tasks[0] = task
	} else {
		self.Tasks = append(self.Tasks, task)
	}
	sort.Sort(self.Tasks)
	return nil
}

func (self *Queue) Add(task *Task) error {
	self.Lock()
	defer self.Unlock()

	ln := len(self.Tasks)
	if ln >= int(self.MaxQueueSize) {
		return fmt.Errorf("queue is full")
	} else {
		self.Tasks = append(self.Tasks, task)
	}
	sort.Sort(self.Tasks)
	return nil
}

func (self *Queue) HasTask(taskId string) bool {
	self.Lock()
	defer self.Unlock()

	for _, task := range self.Tasks {
		if task.TaskId == taskId {
			return true
		}
	}
	return false
}

func (self *Queue) GetTask(taskId string) *Task {
	self.Lock()
	defer self.Unlock()

	for _, task := range self.Tasks {
		if task.TaskId == taskId {
			return task
		}
	}
	return nil
}
