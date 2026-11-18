package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type repo struct {
	FullName string `json:"full_name"`
	Fork     bool   `json:"fork"`
}

type client struct {
	base  string
	token string
	http  *http.Client
}

func (c *client) get(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s: %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *client) repos(visibility string) ([]repo, error) {
	var all []repo
	for page := 1; ; page++ {
		var batch []repo
		path := fmt.Sprintf("/user/repos?affiliation=owner&visibility=%s&per_page=100&page=%d", visibility, page)
		if err := c.get(path, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			return all, nil
		}
	}
}

func (c *client) languages(fullName string) (map[string]int64, error) {
	var langs map[string]int64
	if err := c.get("/repos/"+fullName+"/languages", &langs); err != nil {
		return nil, err
	}
	return langs, nil
}

type entry struct {
	Name  string
	Bytes int64
	Pct   float64
}

func aggregate(totals map[string]int64, exclude map[string]bool, limit int) []entry {
	var list []entry
	var sum int64
	for name, size := range totals {
		if exclude[strings.ToLower(name)] || size <= 0 {
			continue
		}
		list = append(list, entry{Name: name, Bytes: size})
		sum += size
	}
	if sum == 0 {
		return nil
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Bytes != list[j].Bytes {
			return list[i].Bytes > list[j].Bytes
		}
		return list[i].Name < list[j].Name
	})
	if limit > 0 && len(list) > limit {
		var rest int64
		for _, e := range list[limit:] {
			rest += e.Bytes
		}
		list = append(list[:limit], entry{Name: "Other", Bytes: rest})
	}
	for i := range list {
		list[i].Pct = float64(list[i].Bytes) * 100 / float64(sum)
	}
	return list
}

func parseExclude(s string) map[string]bool {
	set := map[string]bool{}
	for _, name := range strings.Split(s, ",") {
		name = strings.ToLower(strings.TrimSpace(name))
		if name != "" {
			set[name] = true
		}
	}
	return set
}

func main() {
	out := flag.String("out", "langcard.svg", "output file")
	exclude := flag.String("exclude", "Jupyter Notebook,HTML,CSS,Makefile", "comma separated languages to ignore")
	limit := flag.Int("limit", 6, "number of languages before grouping into Other")
	visibility := flag.String("visibility", "all", "all, public or private")
	title := flag.String("title", "Most used languages", "card title")
	flag.Parse()

	token := os.Getenv("GH_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "GH_TOKEN is not set")
		os.Exit(1)
	}
	base := os.Getenv("GITHUB_API_URL")
	if base == "" {
		base = "https://api.github.com"
	}

	c := &client{base: strings.TrimRight(base, "/"), token: token, http: &http.Client{Timeout: 30 * time.Second}}

	repos, err := c.repos(*visibility)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	totals := map[string]int64{}
	for _, r := range repos {
		if r.Fork {
			continue
		}
		langs, err := c.languages(r.FullName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for name, size := range langs {
			totals[name] += size
		}
	}

	entries := aggregate(totals, parseExclude(*exclude), *limit)
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no language data found")
		os.Exit(1)
	}

	if err := os.WriteFile(*out, []byte(render(*title, entries)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
