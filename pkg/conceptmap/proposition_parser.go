package conceptmap

import (
	"fmt"
	"strings"
	"unicode"
)

type PropositionParser struct {
}

func (p *PropositionParser) Parse(s string, propositions *PropositionList, concepts *[]*Concept) error {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")

	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		err := parseProposition(line, propositions, concepts)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseProposition(s string, propositions *PropositionList, concepts *[]*Concept) error {
	trimmed := strings.TrimSpace(s)
	words := strings.Fields(trimmed)
	state := 1 // 1: parsing left concept, 2: parsing predicate, 3 parsing right concept
	leftWords := []string{}
	predicateWords := []string{}
	rightWords := []string{}

	for _, word := range words {
		switch state {
		case 1:
			if startsWithLowerCase(word) {
				predicateWords = append(predicateWords, word)
				state = state + 1
			} else {
				leftWords = append(leftWords, word)
			}

		case 2:
			if startsWithLowerCase(word) {
				predicateWords = append(predicateWords, word)
			} else {
				rightWords = append(rightWords, word)
				state = state + 1
			}

		case 3:
			if startsWithLowerCase(word) {
				return fmt.Errorf("encountered unexpected lower case word '%s' outside of predicate", word)
			}
			rightWords = append(rightWords, word)
		}
	}

	if len(leftWords) == 0 {
		return fmt.Errorf("could not find left concept in proposition '%s'", s)
	}

	if len(rightWords) == 0 {
		return fmt.Errorf("could not find right concept in proposition '%s'", s)
	}

	if len(predicateWords) == 0 {
		return fmt.Errorf("could not find predicate in proposition '%s'", s)
	}

	proposition := &Proposition{
		Predicate: Predicate(strings.Join(predicateWords, " ")),
	}

	// Check if we have already parsed either the left or right concepts from another
	// proposition before we create a new one
	leftConceptLabel := strings.Join(leftWords, " ")
	rightConceptLabel := strings.Join(rightWords, " ")

	for i, concept := range *concepts {
		if concept.Label == leftConceptLabel {
			proposition.Left = (*concepts)[i]
		}
		if concept.Label == rightConceptLabel {
			proposition.Right = (*concepts)[i]
		}
	}

	// If the left or right concepts haven't already been created then create them now
	// assign them to the proposition and also store them in the concept list for the map
	if proposition.Left == nil {
		proposition.Left = &Concept{Label: leftConceptLabel}
		*concepts = append(*concepts, proposition.Left)
	}

	if proposition.Right == nil {
		proposition.Right = &Concept{Label: rightConceptLabel}
		*concepts = append(*concepts, proposition.Right)
	}

	*propositions = append(*propositions, proposition)

	return nil
}

func startsWithLowerCase(s string) bool {
	output := false

	// for..range over string automatically converts to unicode runes
	// required for unicode.IsLower()
	for _, c := range s {
		output = unicode.IsLower(c)
		break
	}

	return output
}
