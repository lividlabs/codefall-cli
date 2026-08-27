package domain

// Report is the ordered outcome of one doctor run. A check that was skipped because an earlier one
// failed is absent from the report rather than present with a "skipped" status — there is nothing
// useful to say about a check that never ran.
type Report struct {
	results []Result
}

// NewReport builds a report from results already in check order.
func NewReport(results ...Result) Report {
	return Report{results: results}
}

// Results returns the results in check order.
func (r Report) Results() []Result {
	return r.results
}

// Section is one category's slice of a report: the checks that ran for it, in the order they ran.
type Section struct {
	Category Category
	Results  []Result
}

// Status is the worst status in the section, because that is what a reader needs from a header line:
// a section is only healthy when every check in it passed.
func (s Section) Status() Status {
	worst := StatusPass

	for _, result := range s.Results {
		if result.Status > worst {
			worst = result.Status
		}
	}

	return worst
}

// Sections groups the results by category, in order of first appearance and with the result order
// inside each one preserved. A category whose checks were all skipped has no section at all.
func (r Report) Sections() []Section {
	var sections []Section

	index := make(map[string]int, len(r.results))

	for _, result := range r.results {
		id := result.Check.Category.ID

		position, seen := index[id]
		if !seen {
			position = len(sections)
			index[id] = position
			sections = append(sections, Section{Category: result.Check.Category})
		}

		sections[position].Results = append(sections[position].Results, result)
	}

	return sections
}

// SectionsWithIssues counts the sections that are not clean, warnings included. This is what the
// summary line reports; it is not the exit status, which Failed decides.
func (r Report) SectionsWithIssues() int {
	issues := 0

	for _, section := range r.Sections() {
		if section.Status() != StatusPass {
			issues++
		}
	}

	return issues
}

// Failed counts the results that failed. Warnings do not count: they never change the exit status.
func (r Report) Failed() int {
	failed := 0

	for _, result := range r.results {
		if result.Status == StatusFail {
			failed++
		}
	}

	return failed
}
