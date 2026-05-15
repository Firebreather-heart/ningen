package domain

// Item represents a standard review item extracted from any data source.
type Item struct {
	ID         string
	Domain     string
	Metadata   string
	SearchText string
	Embedding  []float32
}
