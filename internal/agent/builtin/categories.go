package builtin

// Category is the routing key for the task tool. It maps to a model + a
// prompt append. Categories are provider-agnostic; the model field is
// resolved through the user's heretic.json provider config.
type Category string

const (
	// CategoryQuick: trivial tasks (single-file, typo fixes).
	CategoryQuick Category = "quick"
	// CategoryUltrabrain: hard logic-heavy tasks. Use sparingly.
	CategoryUltrabrain Category = "ultrabrain"
	// CategoryDeep: goal-oriented autonomous problem-solving.
	CategoryDeep Category = "deep"
	// CategoryArtistry: creative, unconventional solutions.
	CategoryArtistry Category = "artistry"
	// CategoryVisualEngineering: frontend, UI/UX, design, animation.
	CategoryVisualEngineering Category = "visual-engineering"
	// CategoryWriting: documentation, prose, technical writing.
	CategoryWriting Category = "writing"
	// CategoryUnspecifiedLow: doesn't fit other categories, low effort.
	CategoryUnspecifiedLow Category = "unspecified-low"
	// CategoryUnspecifiedHigh: doesn't fit other categories, high effort.
	CategoryUnspecifiedHigh Category = "unspecified-high"
)

// AllCategories returns the 8 default categories.
func AllCategories() []Category {
	return []Category{
		CategoryQuick, CategoryUltrabrain, CategoryDeep, CategoryArtistry,
		CategoryVisualEngineering, CategoryWriting,
		CategoryUnspecifiedLow, CategoryUnspecifiedHigh,
	}
}

// CategoryConfig is the model + prompt + description for one category.
type CategoryConfig struct {
	// Model is the provider/model identifier (e.g. "openai/gpt-5.4-mini").
	// Users can override this in heretic.json.
	Model string
	// Variant is the model variant (e.g. "max", "high", "xhigh").
	Variant string
}

// CategoryMeta bundles a category's config with its display metadata.
type CategoryMeta struct {
	Name        Category
	Display     string
	Description string
	Model       string
	Variant     string
}

// DefaultCategoryMeta returns the 8 default categories with their
// provider-agnostic defaults. These match OmO's built-in categories
// but are model-agnostic — users override the model field in
// heretic.json. The defaults are the same as OmO's.
func DefaultCategoryMeta() []CategoryMeta {
	return []CategoryMeta{
		{
			Name:        CategoryQuick,
			Display:     "Quick",
			Description: "Trivial tasks - single file changes, typo fixes, simple modifications",
			Model:       "small",
		},
		{
			Name:        CategoryUltrabrain,
			Display:     "Ultrabrain",
			Description: "Use ONLY for genuinely hard, logic-heavy tasks. Give clear goals only, not step-by-step instructions.",
			Model:       "large",
			Variant:     "max",
		},
		{
			Name:        CategoryDeep,
			Display:     "Deep",
			Description: "Goal-oriented autonomous problem-solving on hairy problems requiring deep research. ONE goal + ONE deliverable per call.",
			Model:       "large",
		},
		{
			Name:        CategoryArtistry,
			Display:     "Artistry",
			Description: "Complex problem-solving with unconventional, creative approaches - beyond standard patterns",
			Model:       "large",
			Variant:     "high",
		},
		{
			Name:        CategoryVisualEngineering,
			Display:     "Visual Engineering",
			Description: "Frontend, UI/UX, design, styling, animation",
			Model:       "large",
			Variant:     "high",
		},
		{
			Name:        CategoryWriting,
			Display:     "Writing",
			Description: "Documentation, prose, technical writing",
			Model:       "large",
		},
		{
			Name:        CategoryUnspecifiedLow,
			Display:     "Unspecified (low)",
			Description: "Tasks that don't fit other categories, low effort required",
			Model:       "large",
		},
		{
			Name:        CategoryUnspecifiedHigh,
			Display:     "Unspecified (high)",
			Description: "Tasks that don't fit other categories, high effort required",
			Model:       "large",
			Variant:     "max",
		},
	}
}

// CategoryRegistry maps category names to their metadata.
type CategoryRegistry struct {
	byName map[Category]CategoryMeta
}

// NewCategoryRegistry returns a registry populated with the defaults.
func NewCategoryRegistry() *CategoryRegistry {
	r := &CategoryRegistry{byName: make(map[Category]CategoryMeta)}
	for _, m := range DefaultCategoryMeta() {
		r.byName[m.Name] = m
	}
	return r
}

// Get returns the metadata for a category.
func (r *CategoryRegistry) Get(c Category) (CategoryMeta, bool) {
	m, ok := r.byName[c]
	return m, ok
}

// GetOrError is like Get but returns an error on miss.
func (r *CategoryRegistry) GetOrError(c Category) (CategoryMeta, error) {
	m, ok := r.Get(c)
	if !ok {
		return CategoryMeta{}, &UnknownCategoryError{Name: c}
	}
	return m, nil
}

// All returns all registered categories.
func (r *CategoryRegistry) All() []CategoryMeta {
	out := make([]CategoryMeta, 0, len(r.byName))
	for _, m := range r.byName {
		out = append(out, m)
	}
	return out
}

// UnknownCategoryError is returned when a category is not in the registry.
type UnknownCategoryError struct {
	Name Category
}

func (e *UnknownCategoryError) Error() string {
	return "builtin: unknown category " + string(e.Name)
}
