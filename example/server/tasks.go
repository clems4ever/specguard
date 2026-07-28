package server

type Task struct {
	ID    string
	Title string
	Done  bool
	Owner string
}

type Tasks struct{}

func (Tasks) Create(owner, title string) (Task, error) { return Task{}, nil }
func (Tasks) Toggle(id string) error                   { return nil }
func (Tasks) Delete(id, actor string) error            { return nil }
