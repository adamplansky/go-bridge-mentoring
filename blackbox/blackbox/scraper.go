package blackbox

// reporter reports results from every scrape target
type reporter interface {
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

func (s *Scraper) RunReporter() chan Result {
	resultCh := make(chan Result)
	go s.r.Report(resultCh)
	return resultCh
}

func (s *Scraper) Scrape(r chan Result) {
	for _, job := range s.jobs {
		j := job
		go j.Do(r)
	}
}
