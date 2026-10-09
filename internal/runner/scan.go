package runner

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/validate"
)

type ScanOptions struct {
	Target            string
	DirectoryWordlist string
	FileWordlist      string
	Workers           int
	Codes             []int
	Extensions        []string
	HTTP              HTTPOptions
	Recursive         bool
	MaxDepth          int
	IncludeDirs       bool
	IncludeFiles      bool
	SkipWildcardCheck bool
}

type scanJob struct {
	base  string
	depth int
	kind  string
	word  string
	ext   string
}

type scanEvent struct {
	result output.Result
	next   string
}

type wordlistLoad struct {
	kind     string
	words    []string
	requests int
	err      error
}

func RunScan(ctx context.Context, opts ScanOptions) ([]output.Result, error) {
	if !validate.URL(opts.Target) {
		return nil, ErrInvalidTarget("url")
	}
	if opts.HTTP.Timeout == 0 {
		opts.HTTP.Timeout = 15 * time.Second
	}
	if opts.HTTP.Method == "" {
		opts.HTTP.Method = http.MethodGet
	}
	if !opts.IncludeDirs && !opts.IncludeFiles {
		opts.IncludeDirs = true
	}
	if opts.MaxDepth < 0 {
		opts.MaxDepth = 0
	}

	codes := StatusSet(opts.Codes)
	client := NewHTTPClient(opts.HTTP)
	if !opts.SkipWildcardCheck {
		output.Info("checking wildcard responses before enumeration...")
		if wildcard, err := wildcardStatus(ctx, client, opts.HTTP, opts.Target, codes); err != nil {
			return nil, err
		} else if wildcard {
			output.Warn("target returns a selected status for random missing paths; use --skip-wildcard-check or narrow --codes")
			return []output.Result{}, nil
		}
		output.Info("wildcard check passed")
	}

	output.Info("starting dir enumeration...")
	if opts.Recursive {
		return runContinuousScan(ctx, client, opts, codes)
	}
	return runRecursiveScan(ctx, client, opts, codes)
}

// scanDispatcher is an unbounded, concurrency-safe FIFO work queue. Workers pull
// jobs with next() in the order they were added; when a request discovers a new
// directory, its child jobs are pushed back with add() so recursion runs in
// parallel with the rest of the scan instead of waiting for the current depth to
// finish. FIFO keeps the wordlist order, so common top-of-list directories are
// probed first (same as a non-recursive scan).
type scanDispatcher struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []scanJob
	head    int // index of the next job to hand out
	pending int
	closed  bool
}

func newScanDispatcher() *scanDispatcher {
	d := &scanDispatcher{}
	d.cond = sync.NewCond(&d.mu)
	return d
}

func (d *scanDispatcher) add(jobs []scanJob) {
	if len(jobs) == 0 {
		return
	}
	d.mu.Lock()
	if !d.closed {
		d.queue = append(d.queue, jobs...)
		d.pending += len(jobs)
	}
	d.mu.Unlock()
	d.cond.Broadcast()
}

func (d *scanDispatcher) next() (scanJob, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for d.head >= len(d.queue) && !d.closed {
		d.cond.Wait()
	}
	if d.head >= len(d.queue) {
		return scanJob{}, false
	}
	job := d.queue[d.head]
	d.queue[d.head] = scanJob{} // drop the reference so it can be collected
	d.head++
	// Reclaim memory once most of the backing array has been consumed, keeping
	// peak usage bounded during long recursive scans.
	if d.head > 1024 && d.head*2 >= len(d.queue) {
		d.queue = append(d.queue[:0], d.queue[d.head:]...)
		d.head = 0
	}
	return job, true
}

// done marks one job complete; when nothing is left in flight the queue closes
// and idle workers wake up to exit.
func (d *scanDispatcher) done() {
	d.mu.Lock()
	d.pending--
	closed := d.pending <= 0
	if closed {
		d.closed = true
	}
	d.mu.Unlock()
	if closed {
		d.cond.Broadcast()
	}
}

// cancel closes the queue early (e.g. on Ctrl-C) so workers drain and exit.
func (d *scanDispatcher) cancel() {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()
	d.cond.Broadcast()
}

func runContinuousScan(ctx context.Context, client *http.Client, opts ScanOptions, codes map[int]struct{}) ([]output.Result, error) {
	workers := opts.Workers
	if workers < 1 {
		workers = 1
	}
	dirWords, fileWords, err := loadScanWordlists(opts)
	if err != nil {
		return nil, err
	}

	d := newScanDispatcher()
	seen := map[string]struct{}{strings.TrimRight(opts.Target, "/") + "/": {}}
	var seenMu sync.Mutex

	initial := buildScanJobs(opts, opts.Target, 0, dirWords, fileWords)
	if len(initial) == 0 {
		return nil, nil
	}
	output.Info("parallel recursion enabled: %d initial requests, up to depth %d", len(initial), opts.MaxDepth)
	progress := output.NewProgress("Progress", len(initial))
	defer progress.Finish()

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			d.cancel()
		case <-stop:
		}
	}()

	// Seed the queue before starting workers so the first requests go out
	// immediately and newly discovered directories are interleaved from then on.
	d.add(initial)

	stats := newScanStats()
	var found []output.Result
	var foundMu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				job, ok := d.next()
				if !ok {
					return
				}
				if ctx.Err() == nil {
					processScanJob(ctx, client, opts, codes, job, d, progress, seen, &seenMu, dirWords, fileWords, &found, &foundMu, stats)
				}
				d.done()
			}
		}()
	}
	wg.Wait()
	output.Info("scan summary: %s", stats.summary())
	return found, nil
}

func processScanJob(ctx context.Context, client *http.Client, opts ScanOptions, codes map[int]struct{}, job scanJob, d *scanDispatcher, progress *output.ProgressTracker, seen map[string]struct{}, seenMu *sync.Mutex, dirWords, fileWords []string, found *[]output.Result, foundMu *sync.Mutex, stats *scanStats) {
	result, discovered, ok := executeScanJob(ctx, client, opts, codes, job, stats)
	progress.Advance()
	if !ok {
		return
	}
	foundMu.Lock()
	*found = append(*found, result)
	foundMu.Unlock()
	output.PrintLiveResult(result)
	if discovered == "" {
		return
	}
	seenMu.Lock()
	if _, dup := seen[discovered]; dup {
		seenMu.Unlock()
		return
	}
	seen[discovered] = struct{}{}
	seenMu.Unlock()
	children := buildScanJobs(opts, discovered, job.depth+1, dirWords, fileWords)
	progress.AddTotal(len(children))
	d.add(children)
}

func loadScanWordlists(opts ScanOptions) (dirWords, fileWords []string, err error) {
	if opts.IncludeDirs {
		dirWords, err = LoadWordlist(opts.DirectoryWordlist)
		if err != nil {
			return nil, nil, err
		}
	}
	if opts.IncludeFiles {
		fileWords, err = LoadWordlist(opts.FileWordlist)
		if err != nil {
			return nil, nil, err
		}
	}
	return dirWords, fileWords, nil
}

func buildScanJobs(opts ScanOptions, base string, depth int, dirWords, fileWords []string) []scanJob {
	var jobs []scanJob
	if opts.IncludeDirs {
		for _, word := range dirWords {
			jobs = append(jobs, scanJob{base: base, depth: depth, kind: "directory", word: word})
		}
	}
	if opts.IncludeFiles {
		for _, word := range fileWords {
			for _, ext := range opts.Extensions {
				jobs = append(jobs, scanJob{base: base, depth: depth, kind: "file", word: word, ext: ext})
			}
		}
	}
	return jobs
}

func runRecursiveScan(ctx context.Context, client *http.Client, opts ScanOptions, codes map[int]struct{}) ([]output.Result, error) {
	workers := opts.Workers
	if workers < 1 {
		workers = 1
	}
	var found []output.Result
	var dirWords []string
	var fileWords []string
	stats := newScanStats()
	seen := map[string]struct{}{strings.TrimRight(opts.Target, "/") + "/": {}}
	bases := []string{opts.Target}

	for depth := 0; len(bases) > 0; depth++ {
		output.Debug("scanning depth %d with %d base URL(s)", depth, len(bases))
		total := len(bases) * len(dirWords)
		if opts.IncludeFiles {
			total += len(bases) * len(fileWords) * len(opts.Extensions)
		}
		if depth == 0 {
			output.Info("dispatching requests at depth %d with %d workers as the wordlist is read", depth, workers)
			total = 0
		} else {
			output.Info("dispatching %d requests at depth %d with %d workers", total, depth, workers)
		}
		progress := output.NewProgress("Progress depth "+itoa(depth), total)
		jobs := make(chan scanJob, workers*2)
		events := make(chan scanEvent, workers)
		loads := make(chan wordlistLoad, 2)
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					result, discovered, ok := executeScanJob(ctx, client, opts, codes, job, stats)
					progress.Advance()
					if !ok {
						continue
					}
					events <- scanEvent{result: result, next: discovered}
				}
			}()
		}

		if depth == 0 {
			go streamInitialJobs(opts, bases, depth, jobs, loads, progress)
		} else {
			go enqueueKnownJobs(opts, bases, depth, dirWords, fileWords, jobs)
		}
		go func() {
			wg.Wait()
			close(events)
		}()

		var next []string
		for event := range events {
			found = append(found, event.result)
			output.PrintLiveResult(event.result)
			if event.next != "" {
				if _, ok := seen[event.next]; ok {
					continue
				}
				seen[event.next] = struct{}{}
				next = append(next, event.next)
			}
		}
		if depth == 0 {
			close(loads)
			knownTotal := 0
			for load := range loads {
				if load.err != nil {
					return found, load.err
				}
				knownTotal += load.requests
				switch load.kind {
				case "directory":
					dirWords = load.words
					output.Info("loaded %d directory words", len(dirWords))
				case "file":
					fileWords = load.words
					output.Info("loaded %d file words", len(fileWords))
				}
			}
			progress.SetTotal(knownTotal)
		}
		progress.Finish()
		if !opts.Recursive || depth >= opts.MaxDepth {
			break
		}
		bases = next
	}
	output.Info("scan summary: %s", stats.summary())
	return found, nil
}

func streamInitialJobs(opts ScanOptions, bases []string, depth int, jobs chan<- scanJob, loads chan<- wordlistLoad, progress *output.ProgressTracker) {
	var producers sync.WaitGroup
	var totals sync.WaitGroup
	totalCh := make(chan int, 2)

	if opts.IncludeDirs {
		producers.Add(1)
		totals.Add(1)
		go func() {
			defer producers.Done()
			defer totals.Done()
			words, requests, err := streamWordlistJobs("directory", opts.DirectoryWordlist, bases, depth, "", jobs)
			totalCh <- requests
			loads <- wordlistLoad{kind: "directory", words: words, requests: requests, err: err}
		}()
	}
	if opts.IncludeFiles {
		producers.Add(1)
		totals.Add(1)
		go func() {
			defer producers.Done()
			defer totals.Done()
			words, requests, err := streamFileWordlistJobs(opts.FileWordlist, bases, depth, opts.Extensions, jobs)
			totalCh <- requests
			loads <- wordlistLoad{kind: "file", words: words, requests: requests, err: err}
		}()
	}

	go func() {
		producers.Wait()
		close(jobs)
	}()
	go func() {
		totals.Wait()
		close(totalCh)
	}()

	knownTotal := 0
	for total := range totalCh {
		knownTotal += total
		progress.SetTotal(knownTotal)
	}
}

func enqueueKnownJobs(opts ScanOptions, bases []string, depth int, dirWords, fileWords []string, jobs chan<- scanJob) {
	var producers sync.WaitGroup
	if opts.IncludeDirs {
		producers.Add(1)
		go func() {
			defer producers.Done()
			for _, base := range bases {
				for _, word := range dirWords {
					jobs <- scanJob{base: base, depth: depth, kind: "directory", word: word}
				}
			}
		}()
	}
	if opts.IncludeFiles {
		producers.Add(1)
		go func() {
			defer producers.Done()
			for _, base := range bases {
				for _, word := range fileWords {
					for _, ext := range opts.Extensions {
						jobs <- scanJob{base: base, depth: depth, kind: "file", word: word, ext: ext}
					}
				}
			}
		}()
	}
	producers.Wait()
	close(jobs)
}

func streamWordlistJobs(kind, path string, bases []string, depth int, ext string, jobs chan<- scanJob) ([]string, int, error) {
	output.Info("reading %s wordlist %s and sending requests immediately...", kind, path)
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var words []string
	requests := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*64), 1024*1024)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		words = append(words, word)
		for _, base := range bases {
			jobs <- scanJob{base: base, depth: depth, kind: kind, word: word, ext: ext}
			requests++
		}
	}
	return words, requests, scanner.Err()
}

func streamFileWordlistJobs(path string, bases []string, depth int, extensions []string, jobs chan<- scanJob) ([]string, int, error) {
	output.Info("reading file wordlist %s and sending requests immediately...", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var words []string
	requests := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*64), 1024*1024)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		words = append(words, word)
		for _, base := range bases {
			for _, ext := range extensions {
				jobs <- scanJob{base: base, depth: depth, kind: "file", word: word, ext: ext}
				requests++
			}
		}
	}
	return words, requests, scanner.Err()
}

func executeScanJob(ctx context.Context, client *http.Client, opts ScanOptions, codes map[int]struct{}, job scanJob, stats *scanStats) (output.Result, string, bool) {
	switch job.kind {
	case "directory":
		url := JoinURL(job.base, job.word, true)
		result, status, ok, err := doHTTPRequest(ctx, client, opts.HTTP, url, codes)
		stats.record(status, ok, err)
		if !ok {
			return output.Result{}, "", false
		}
		result.Type = "directory"
		result.Target = opts.Target
		result.URL = url
		result.Path = resultPath(url)
		discovered := ""
		if opts.Recursive && job.depth < opts.MaxDepth {
			discovered = url
		}
		return result, discovered, true
	case "file":
		name := strings.Trim(strings.TrimSpace(job.word), ".")
		if name == "" {
			return output.Result{}, "", false
		}
		cleanExt := strings.Trim(strings.TrimSpace(job.ext), ".")
		if cleanExt == "" {
			return output.Result{}, "", false
		}
		url := JoinURL(job.base, name+"."+cleanExt, false)
		result, status, ok, err := doHTTPRequest(ctx, client, opts.HTTP, url, codes)
		stats.record(status, ok, err)
		if !ok {
			return output.Result{}, "", false
		}
		result.Type = "file"
		result.Target = opts.Target
		result.URL = url
		result.Path = resultPath(url)
		return result, "", true
	default:
		return output.Result{}, "", false
	}
}

// scanStats aggregates request outcomes so a scan can report why it did or did
// not find anything, even when every request failed or was filtered out.
type scanStats struct {
	mu      sync.Mutex
	total   int
	errors  int
	matched int
	status  map[int]int
}

func newScanStats() *scanStats {
	return &scanStats{status: map[int]int{}}
}

func (s *scanStats) record(status int, matched bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total++
	if err != nil {
		s.errors++
		return
	}
	s.status[status]++
	if matched {
		s.matched++
	}
}

func (s *scanStats) summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	parts := []string{fmt.Sprintf("%d requests", s.total), fmt.Sprintf("%d matched", s.matched)}
	if s.errors > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", s.errors))
	}
	// Top non-matching status codes, most frequent first, to hint at WAFs (403),
	// redirects not in --codes, soft 404s, etc.
	type kv struct {
		code, n int
	}
	var codes []kv
	for code, n := range s.status {
		codes = append(codes, kv{code, n})
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i].n > codes[j].n })
	for i, c := range codes {
		if i >= 4 {
			break
		}
		parts = append(parts, fmt.Sprintf("%dx%d", c.n, c.code))
	}
	return strings.Join(parts, ", ")
}

func resultPath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Path == "" {
		return rawURL
	}
	return parsed.EscapedPath()
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
