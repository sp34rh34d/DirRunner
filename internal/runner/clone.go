package runner

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"dirrunner/internal/output"
)

type CloneOptions struct {
	Target      string
	ProjectName string
	Workers     int
	HTTP        HTTPOptions
}

// RunClone downloads a site breadth-first, following same-domain links and
// assets, and writes every fetched resource under the project directory. It
// mirrors WebRunner's clone module.
func RunClone(ctx context.Context, opts CloneOptions) ([]output.Result, error) {
	if !ValidURL(opts.Target) {
		return nil, fmt.Errorf("invalid target URL: %s", opts.Target)
	}
	project := opts.ProjectName
	if project == "" {
		project = "clone-" + RandomToken()
	}
	start, err := url.Parse(opts.Target)
	if err != nil {
		return nil, err
	}
	domain := start.Host

	workers := opts.Workers
	if workers < 1 {
		workers = 1
	}
	client := NewHTTPClient(opts.HTTP)
	output.Info("cloning %s into %s/ with %d workers", opts.Target, project, workers)

	var mu sync.Mutex
	visited := map[string]struct{}{opts.Target: {}}
	var all []output.Result

	// sem bounds the number of in-flight requests; one goroutine is spawned per
	// discovered same-domain URL and blocks on sem until a slot is free.
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	var crawl func(current string)
	crawl = func(current string) {
		defer wg.Done()
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			return
		}

		status, headers, body, err := FetchResponse(ctx, client, opts.HTTP, "GET", current, nil)
		if err != nil {
			output.Debug("error fetching %s: %v", current, err)
			return
		}
		savedPath, err := saveClonedResource(project, current, body)
		if err != nil {
			output.Warn("could not save %s: %v", current, err)
			return
		}
		output.Debug("saved %s", savedPath)
		result := output.Result{
			Type:   "clone",
			Target: current,
			Method: "GET",
			URL:    current,
			Status: status,
			Size:   int64(len(body)),
			Path:   savedPath,
		}
		mu.Lock()
		all = append(all, result)
		mu.Unlock()
		output.PrintLiveResult(result)

		// Only parse HTML documents for more links.
		if !strings.Contains(strings.ToLower(headers.Get("Content-Type")), "text/html") {
			return
		}
		text := string(body)
		refs := append(extractMatches(reHrefAttr, text), extractMatches(reLinkHref, text)...)
		refs = append(refs, extractMatches(reScriptSrc, text)...)
		refs = append(refs, extractMatches(reImgSrc, text)...)
		for _, ref := range refs {
			abs := resolveURL(current, ref)
			if abs == "" {
				continue
			}
			u, err := url.Parse(abs)
			if err != nil || u.Host != domain {
				continue
			}
			mu.Lock()
			if _, done := visited[abs]; !done {
				visited[abs] = struct{}{}
				wg.Add(1)
				go crawl(abs)
			}
			mu.Unlock()
		}
	}

	wg.Add(1)
	go crawl(opts.Target)
	wg.Wait()

	output.Info("clone complete: %d resources saved under %s/", len(all), project)
	return all, nil
}

// saveClonedResource writes body to a file derived from rawURL under baseDir,
// guarding against path-traversal in attacker-controlled URLs.
func saveClonedResource(baseDir, rawURL string, body []byte) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	p := u.Path
	switch {
	case p == "" || p == "/":
		p = "/index.html"
	case strings.HasSuffix(p, "/"):
		p += "index.html"
	case path.Ext(p) == "":
		p += ".html"
	}

	target := safeJoin(baseDir, u.Host, p)
	if target == "" {
		return "", fmt.Errorf("refusing unsafe path for %s", rawURL)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(target, body, 0o644); err != nil {
		return "", err
	}
	return target, nil
}

// safeJoin builds baseDir/host/cleanPath while dropping "." and ".." components
// so a remote URL can never write outside baseDir.
func safeJoin(baseDir, host, p string) string {
	host = strings.NewReplacer("/", "_", "\\", "_", "..", "_").Replace(host)
	cleaned := path.Clean("/" + p)
	parts := []string{baseDir, host}
	for _, part := range strings.Split(cleaned, "/") {
		if part == "" || part == "." || part == ".." {
			continue
		}
		parts = append(parts, part)
	}
	return filepath.Join(parts...)
}
