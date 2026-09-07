package item

const delim = "---"

type Item struct {
	Front Frontmatter
	Body  string
	Path  string
}
