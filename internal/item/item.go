package item

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

type Item struct {
	Front Frontmatter
	Body  string
	Path  string
}
