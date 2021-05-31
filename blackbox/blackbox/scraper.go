package blackbox

type reporter interface {
	// Report reports all responses from given channel.
	Report(resp <-chan Result)
}

type Scraper struct {
	jobs []Job
	r    reporter
}

func NewScraper(jobs []Job, r reporter) (*Scraper, error) {
	for _, job := range jobs {
		if err := job.IsValid(); err != nil {
			return nil, err
		}
	}

	return &Scraper{
		jobs: jobs,
		r:    r,
	}, nil
}

// RunReporter crates new report channel and spin up new go routine with Report method
// with this newly created channel. Every scraped job needs to report result into this
// newly created report channel.
func (s *Scraper) RunReporter() chan Result {
	resultCh := make(chan Result)
	go s.r.Report(resultCh)
	return resultCh
}

// Scrape starts to scrape all jobs. Every job is done in separate go routine.
func (s *Scraper) Scrape(r chan Result) {
	for _, job := range s.jobs {
		j := job
		go j.Do(r)
	}
}
