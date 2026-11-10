package corpus

import (
	"fmt"

	"github.com/incurl/sumate/internal/annotate"
	"github.com/incurl/sumate/internal/model"
)

// Pipeline assembles the annotated verse set from the raw corpus.
type Pipeline struct {
	corpus *Corpus
	engine *annotate.Engine
}

// NewPipeline returns a pipeline over the given corpus and annotation engine.
func NewPipeline(c *Corpus, e *annotate.Engine) *Pipeline {
	return &Pipeline{corpus: c, engine: e}
}

// Build produces every minimal matching unit with its full annotation.
func (p *Pipeline) Build() ([]model.Verse, error) {
	var verses []model.Verse
	for _, poem := range p.corpus.Poems {
		bg := DeriveHistory(poem)
		poem.Era, poem.Year, poem.Stage, poem.Place = bg.Era, bg.Year, bg.Stage, bg.Place

		units := SplitUnits(poem)
		for _, u := range units {
			v := model.Verse{
				ID:         fmt.Sprintf("%s-%03d", poem.ID, u.Order),
				PoemID:     poem.ID,
				Tune:       poem.Tune,
				Title:      poem.Title,
				Text:       u.Text,
				Kind:       u.Kind,
				Order:      u.Order,
				Stanza:     u.Stanza,
				Background: bg,
			}
			v.Annotation = p.engine.Annotate(&v, poem)
			verses = append(verses, v)
		}
	}
	return verses, nil
}
