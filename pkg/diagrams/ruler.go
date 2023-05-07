package diagrams

import "oss.terrastruct.com/d2/lib/textmeasure"

type D2RulerFactory func() (*textmeasure.Ruler, error)

func defaultRulerFactory() (*textmeasure.Ruler, error) {
	return textmeasure.NewRuler()
}
