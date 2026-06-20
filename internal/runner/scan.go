package runner

import (
	"bufio"
	"context"
	"net/http"
	"net/url"
	"os"
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
	return runRecursiveScan(ctx, client, opts, codes)
}

func runRecursiveScan(ctx context.Context, client *http.Client, opts ScanOptions, codes map[int]struct{}) ([]output.Result, error) {
	workers := opts.Workers
	if workers < 1 {
		workers = 1
	}
	var found []output.Result
	var dirWords []string
	var fileWords []string
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
					result, discovered, ok := executeScanJob(ctx, client, opts, codes, job)
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

func executeScanJob(ctx context.Context, client *http.Client, opts ScanOptions, codes map[int]struct{}, job scanJob) (output.Result, string, bool) {
	switch job.kind {
	case "directory":
		url := JoinURL(job.base, job.word, true)
		result, ok := executeHTTPRequest(ctx, client, opts.HTTP, url, codes)
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
		result, ok := executeHTTPRequest(ctx, client, opts.HTTP, url, codes)
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
