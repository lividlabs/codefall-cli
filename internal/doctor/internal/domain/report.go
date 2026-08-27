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
