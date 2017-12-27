package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type Section struct {
	Title string
	Body  string
}

func (this Section) String() string {
	b := bytes.NewBuffer(nil)
	fmt.Fprintln(b, this.Title+":")

	s := bufio.NewScanner(strings.NewReader(this.Body))
	for s.Scan() {
		t := s.Text()
		fmt.Fprintln(b, "|"+t)
	}

	return b.String()
}

func parseSections(commentBlock string) []Section {

	var (
		sections []Section     = make([]Section, 0)
		section  *Section      // Leave nil until a new section is identified.
		body     *bytes.Buffer = bytes.NewBuffer(nil)
	)

	scnr := bufio.NewScanner(strings.NewReader(commentBlock))
	for scnr.Scan() {
		line := scnr.Text()
		line = strings.TrimSpace(line)
		line_ := strings.ToLower(line)

		// The most basic criteria for finding a section.
		if strings.HasPrefix(line_, "openapi") && strings.HasSuffix(line_, ":") {

			// A new tag means the start of a new section.
			// A new section means the end of a previous section.

			if section != nil {
				// A new section was previously detected.
				// Build it up and spit it out.
				section.Body = body.String()

				sections = append(sections, cleanupSection(*section))
			}
			section = new(Section)
			section.Title = strings.TrimSuffix(line, ":")
			body = bytes.NewBuffer(nil)
		} else {
			fmt.Fprintln(body, line)
		}
	}

	return sections
}

func cleanupSection(section Section) Section {
	var out Section

	var (
		spacesRemoved bool = true // gotta get the loop going.
		tabsRemoved   bool
	)

	for spacesRemoved || tabsRemoved {

		// Remove uniform spaces.
		body := trimPrefixMultiline(section.Body, " ")
		spacesRemoved = len(body) != len(section.Body)
		section.Body = body

		// Remove uniform tabs.
		body = trimPrefixMultiline(section.Body, "\t")
		tabsRemoved = len(body) != len(section.Body)
		section.Body = body
	}

}

func trimPrefixMultiline(s string, prefix string) string {
	scnr := bufio.NewScanner(strings.NewReader(s))
	out := bytes.NewBuffer(nil)
	for scnr.Scan() {
		line := scnr.Text()

		// Empty lines get a pass.
		if len(strings.TrimSpace(line)) == 0 {
			fmt.Fprintln(out)
			continue
		}

		// Every line must have the same prefix.
		if strings.Index(line, prefix) != 0 {
			return s
		}

		// If everything went well, write the trimmed line to the output buffer.
		fmt.Fprintln(out, strings.TrimPrefix(s, prefix))
	}

	return out.String()
}
