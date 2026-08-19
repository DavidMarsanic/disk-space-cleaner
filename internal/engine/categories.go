package engine

// categoryDef is one well-known, documented-safe cache/junk location.
// Paths are resolved lazily (not at package init) since some depend on
// os.UserHomeDir, which can fail in unusual environments.
type categoryDef struct {
	ID          string
	Name        string
	Description string
	resolve     func() []string
}

// Category is what a scan reports about one categoryDef that actually
// exists on this machine — categories that resolve to nothing (the tool
// isn't installed, the cache is already empty) are simply omitted rather
// than shown at zero, since a page full of "0 bytes" rows is just noise.
type Category struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Paths       []string `json:"paths"`
	Bytes       int64    `json:"bytes"`
}
