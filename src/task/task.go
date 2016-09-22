package task

var standardMetadata = []string{
	"name",
	"description",
	"author",
	"date",
	"version",
}

// Task represents a parsed task file
type Task struct {
	Metadata map[string]string

	Concurrent int32
	Template   string
	Prompt     string

	Inventory string
	Devices   []string

	currentBlock        string
	DefaultCommandBlock string
	Commands            map[string]*CommandBlock
}

// CommandBlock contains all the settings for a block of commands
type CommandBlock struct {
	Name     string
	Type     string
	Commands []string
}

func (t *Task) GetMetadata(s string) string {
	data, _ := t.Metadata[s]
	return data
}

func (t *Task) GetAllMetadata() map[string]string {
	return t.Metadata
}

func (t *Task) SetUserData(k, v string) {
	// User data is prefixed with an underscore internally
	// to separate it from internal data
	t.Metadata["_"+k] = v
}
