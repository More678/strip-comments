// Command strip-comments removes every comment from the source files under
// the given paths (the current folder by default).
//
// Nothing is written unless -w is passed: a plain run only reports what it
// would change.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/More678/strip-comments/internal/glossary"
	"github.com/More678/strip-comments/internal/lang"
	"github.com/More678/strip-comments/internal/scan"
	"github.com/More678/strip-comments/internal/textio"
	"github.com/More678/strip-comments/internal/udiff"
	"github.com/More678/strip-comments/internal/verify"
	"github.com/More678/strip-comments/internal/walk"
)

// version is overridden at build time:
// go build -ldflags "-X main.version=v1.2.3"
var version = ""

// buildVersion reports the release version. A binary from build.sh has it
// stamped in; one from `go install pkg@v1.2.3` has to read it back out of the
// module metadata, which is the only place the tag survives there.
func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" &&
		info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

// Exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

type config struct {
	write    bool
	noBackup bool
	stdout   bool
	force    bool
	risky    bool
	quiet    bool
	lang     *lang.Language
	workers  int
	// showTerms lists each term replacement under its file. It is the main
	// point of a dry run with -replace, so it is on whenever -replace is.
	showTerms bool
	// keepComments mirrors scan.Options.KeepComments, so the report can leave
	// out a comment count that is always zero.
	keepComments bool
	// allowSecrets mirrors walk.Options.AllowSecrets for the -stdout path,
	// which does not go through walk.
	allowSecrets bool
	// noVerify skips the structural check that runs before a file is
	// reported as changed.
	noVerify bool
	// diff prints a unified diff per changed file instead of a one-line tally.
	diff bool
	// check makes a run that would change anything exit non-zero, for CI.
	check bool
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	fs := flag.NewFlagSet("strip-comments", flag.ExitOnError)
	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintf(out, "usage: strip-comments [flags] [path ...]\n\n")
		fmt.Fprintf(out, "Removes all comments from the source files under each path\n")
		fmt.Fprintf(out, "(default \".\"). Without -w nothing is written: the run only\n")
		fmt.Fprintf(out, "reports the files it would change.\n\n")
		fs.PrintDefaults()
	}

	var (
		write     = fs.Bool("w", false, "rewrite the files in place (default: dry run)")
		writeLong = fs.Bool("write", false, "alias for -w")
		backup    = fs.Bool("backup", true, "keep <file>.bak before rewriting; -backup=false opts out")
		restore   = fs.Bool("restore", false, "restore named source files from <file>.bak (preview unless -w)")
		toStdout  = fs.Bool("stdout", false, "print one stripped file to stdout, write nothing")
		langName  = fs.String("lang", "", "force this language (see -list)")
		extList   = fs.String("ext", "", "only these extensions, e.g. \".go,.rs\"")
		exclude   = fs.String("exclude", "", "extra names or globs to skip, comma separated")
		keepDoc   = fs.Bool("keep-doc", false, "keep documentation comments (///, /** */)")
		keepBlank = fs.Bool("keep-blank", false, "leave an empty line instead of deleting it")
		keepLic   = fs.Bool("keep-license", false, "keep the leading file comment")
		force     = fs.Bool("force", false, "strip even when the file looks unsafe")
		risky     = fs.Bool("risky", false, "include languages with heuristic support")
		jobs      = fs.Int("j", runtime.NumCPU(), "number of parallel workers")
		quiet     = fs.Bool("q", false, "only report errors")
		list      = fs.Bool("list", false, "list the supported languages and exit")
		showVer   = fs.Bool("version", false, "print the version and exit")
		maxSize   = fs.Int64("max-size", walk.DefaultMaxSize, "skip files larger than this many bytes")

		replace   = fs.Bool("replace", false, "replace terminology from the glossary in code and comments")
		termsFile = fs.String("terms", "", "term table file (TSV or .json) replacing the built-in one")
		termsAdd  = fs.String("terms-add", "", "term table file taking precedence over the built-in one")
		termsCmt  = fs.Bool("terms-in-comments", false, "replace terms only inside comments")
		keepCmt   = fs.Bool("keep-comments", false, "replace terms without removing any comment")
		listTerms = fs.Bool("list-terms", false, "print the term table in effect and exit")
		allowSec  = fs.Bool("allow-secrets", false, "process files that look like credentials (.env, *.pem, secrets.yml)")
		noVerify  = fs.Bool("no-verify", false, "skip the structural check that runs before writing")
		showDiff  = fs.Bool("diff", false, "print a unified diff of every change, write nothing")
		checkOnly = fs.Bool("check", false, "exit non-zero if any file would change, write nothing")
	)
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}

	if *restore {
		// Restoration copies bytes, so transformation and traversal flags cannot apply.
		invalid := ""
		fs.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "restore", "w", "write", "q", "allow-secrets":
			default:
				invalid = f.Name
			}
		})
		if invalid != "" {
			fmt.Fprintf(os.Stderr, "-restore cannot be combined with -%s\n", invalid)
			return exitUsage
		}
		if fs.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "-restore needs explicit source file paths (not directories or .bak files)")
			return exitUsage
		}
		return runRestore(fs.Args(), *write || *writeLong, *quiet, *allowSec)
	}

	switch {
	case *showVer:
		fmt.Printf("strip-comments %s\n", buildVersion())
		return exitOK
	case *list:
		printLanguages()
		return exitOK
	case *listTerms:
		rules, err := termRules(*termsFile, *termsAdd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitUsage
		}
		printTerms(rules)
		return exitOK
	}

	cfg := config{
		write:    *write || *writeLong,
		noBackup: !*backup,
		stdout:   *toStdout,
		force:    *force,
		risky:    *risky,
		quiet:    *quiet,
		workers:  *jobs,

		allowSecrets: *allowSec,
		noVerify:     *noVerify,
		diff:         *showDiff,
		check:        *checkOnly,
	}
	// -check and -diff only report, so asking to write as well is a
	// contradiction worth naming rather than silently resolving.
	if (*checkOnly || *showDiff) && (*write || *writeLong) {
		fmt.Fprintln(os.Stderr, "-check and -diff never write; drop -w")
		return exitUsage
	}
	if cfg.workers < 1 {
		cfg.workers = 1
	}
	if *langName != "" {
		cfg.lang = lang.ByName(*langName)
		if cfg.lang == nil {
			fmt.Fprintf(os.Stderr, "unknown language %q, try -list\n", *langName)
			return exitUsage
		}
	}
	opt := scan.Options{KeepDoc: *keepDoc, KeepBlank: *keepBlank,
		KeepLicense: *keepLic, Force: *force}

	// A term file without -replace is a mistake worth naming: the run would
	// silently ignore the table the user went to the trouble of writing.
	if !*replace && (*termsFile != "" || *termsAdd != "" || *termsCmt || *keepCmt) {
		fmt.Fprintln(os.Stderr,
			"term flags need -replace; nothing would be replaced without it")
		return exitUsage
	}
	if *replace {
		rules, err := termRules(*termsFile, *termsAdd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitUsage
		}
		opt.Glossary = glossary.NewTable(rules)
		opt.TermsInComments = *termsCmt
		opt.KeepComments = *keepCmt
		cfg.showTerms = true
		cfg.keepComments = *keepCmt
	}

	roots := fs.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	if cfg.stdout {
		if len(roots) != 1 {
			fmt.Fprintln(os.Stderr, "-stdout needs exactly one file")
			return exitUsage
		}
		return runStdout(roots[0], cfg, opt)
	}

	self, _ := os.Executable()
	files, skips, errs := walk.Walk(roots, walk.Options{
		Lang:     cfg.lang,
		Exts:     splitList(*extList),
		Exclude:  splitList(*exclude),
		Risky:    cfg.risky,
		MaxSize:  *maxSize,
		SelfPath: self,

		AllowSecrets: *allowSec,
	})
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	results := process(files, cfg, opt)
	return report(results, skips, len(errs), cfg)
}

// termRules assembles the rule set: -terms replaces the built-in table,
// -terms-add takes precedence over it.
//
// Rules are matched longest-first and, at equal length, in slice order, so a
// user rule has to sit *before* the built-in one to win. That is why an added
// file is prepended even though the flag is called -terms-add.
func termRules(replaceFile, addFile string) ([]glossary.Rule, error) {
	base := glossary.Default
	if replaceFile != "" {
		loaded, err := glossary.Load(replaceFile)
		if err != nil {
			return nil, err
		}
		base = loaded
	}
	if addFile == "" {
		return base, nil
	}
	added, err := glossary.Load(addFile)
	if err != nil {
		return nil, err
	}
	return append(added, base...), nil
}

// printTerms lists the rules in match order.
func printTerms(rules []glossary.Rule) {
	tb := glossary.NewTable(rules)
	for _, r := range tb.Rules() {
		flags := ""
		switch {
		case r.Word && r.CI:
			flags = "\tword,ci"
		case r.Word:
			flags = "\tword"
		case r.CI:
			flags = "\tci"
		case glossary.HasASCIIEdges(r.From):
			// An empty column would be read back with the inferred word,ci,
			// so a rule that wants neither has to say so.
			flags = "\tnoword,nocase"
		}
		fmt.Printf("%s\t%s%s\n", r.From, r.To, flags)
	}
}

// counts renders the per-file tally. With -keep-comments the comment count is
// always zero, so printing "-0 comments" would just be noise.
func counts(r result, cfg config) string {
	var parts []string
	if !cfg.keepComments {
		parts = append(parts, fmt.Sprintf("-%d comments", r.comments),
			fmt.Sprintf("-%d lines", r.lines))
	}
	if cfg.showTerms {
		parts = append(parts, fmt.Sprintf("%d terms", len(r.terms)))
	}
	return strings.Join(parts, ", ")
}

// printHits lists each replacement under its file, newest-first by line.
// A dry run with -replace is the main way to review a term table, so the
// before and after text are both shown.
func printHits(hits []glossary.Hit) {
	for _, h := range hits {
		where := "code"
		if h.InComment {
			where = "comment"
		}
		fmt.Printf("      line %d  %s -> %s  (%s)\n", h.Line, h.From, h.To, where)
	}
}

// result is what happened to one file.
type result struct {
	file     walk.File
	changed  bool
	comments int
	lines    int
	terms    []glossary.Hit
	skip     string
	err      error
	// diff holds the unified diff, filled in only under -diff so an ordinary
	// run does not keep two copies of every file in memory.
	diff string
}

// process strips every file through a worker pool. Results stay in walk
// order, so the report is deterministic whatever the workers do.
func process(files []walk.File, cfg config, opt scan.Options) []result {
	results := make([]result, len(files))
	if len(files) == 0 {
		return results
	}
	workers := cfg.workers
	if workers > len(files) {
		workers = len(files)
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				results[idx] = stripOne(files[idx], cfg, opt)
			}
		}()
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// stripOne reads, strips and (in write mode) rewrites a single file.
func stripOne(f walk.File, cfg config, opt scan.Options) result {
	res := result{file: f}
	raw, err := os.ReadFile(f.Path)
	if err != nil {
		res.err = err
		return res
	}
	if textio.IsBinary(raw) {
		res.skip = "binary file"
		return res
	}
	text, format := textio.Decode(raw)
	sr := scan.Strip(text, f.Lang, opt)
	if sr.Unsafe && !cfg.force {
		res.skip = sr.Reason
		return res
	}
	res.terms = sr.Terms
	if sr.Comments == 0 && len(sr.Terms) == 0 {
		// Nothing was stripped or replaced, so leave the bytes alone.
		// Re-encoding would still rewrite a file whose line endings are mixed,
		// and normalising line endings is not this tool's job. Terms count as
		// a change: with -keep-comments they are the only change there is.
		return res
	}
	// Verify before anything is reported as changed, so a dry run shows the
	// same SKIP the write would have produced.
	if !cfg.noVerify {
		if err := verify.Check(text, sr.Output, f.Lang); err != nil {
			res.skip = "verification failed: " + err.Error()
			return res
		}
	}
	out := textio.Encode(sr.Output, format)
	if bytes.Equal(out, raw) {
		return res
	}
	res.changed = true
	res.comments = sr.Comments
	res.lines = sr.RemovedLines
	if cfg.diff {
		// Diff the decoded text, not the raw bytes: a CRLF file would
		// otherwise show a stray \r at the end of every line.
		res.diff = udiff.Unified(displayPath(f.Path), text, sr.Output)
	}
	if !cfg.write {
		return res
	}
	if !cfg.noBackup {
		if err := textio.WriteExclusive(f.Path+".bak", raw, f.Mode.Perm()); err != nil {
			res.err = fmt.Errorf("backup failed; source unchanged: %w", err)
			return res
		}
	}
	if err := textio.WriteAtomic(f.Path, out); err != nil {
		res.err = err
	}
	return res
}

// report prints one line per changed file, then the summary, and returns the
// exit code.
func report(results []result, skips []walk.Skip, walkErrs int, cfg config) int {
	var changed, comments, lines, terms, errCount int
	skipped := len(skips)

	for _, r := range results {
		switch {
		case r.err != nil:
			errCount++
			fmt.Fprintf(os.Stderr, "%s: %v\n", r.file.Path, r.err)
		case r.skip != "":
			skipped++
			if !cfg.quiet {
				fmt.Printf("SKIP %s: %s\n", displayPath(r.file.Path), r.skip)
			}
		case r.changed:
			changed++
			comments += r.comments
			lines += r.lines
			terms += len(r.terms)
			if !cfg.quiet {
				fmt.Printf("  %s  %s  %s\n", displayPath(r.file.Path),
					r.file.Lang.Name, counts(r, cfg))
				if cfg.showTerms {
					printHits(r.terms)
				}
			}
			// The diff is the point of -diff, so -q does not hide it.
			if cfg.diff {
				fmt.Print(r.diff)
			}
			// -check names the files on stderr, where CI logs look first,
			// and regardless of -q.
			if cfg.check {
				fmt.Fprintf(os.Stderr, "would change: %s\n", displayPath(r.file.Path))
			}
		}
	}
	if !cfg.quiet {
		for _, s := range skips {
			fmt.Printf("SKIP %s: %s\n", displayPath(s.Path), s.Reason)
		}
		fmt.Printf("\n%d files scanned, %d changed, ", len(results), changed)
		if !cfg.keepComments {
			fmt.Printf("%d comments removed, %d lines removed, ", comments, lines)
		}
		if cfg.showTerms {
			fmt.Printf("%d terms replaced, ", terms)
		}
		fmt.Printf("%d skipped, %d errors\n", skipped, errCount+walkErrs)
		if !cfg.write && !cfg.check && changed > 0 {
			fmt.Println("dry run: re-run with -w to apply")
		}
	}
	if errCount+walkErrs > 0 {
		return exitError
	}
	// -check fails the build when anything would change, the same contract
	// as `gofmt -l` piped into a non-empty test.
	if cfg.check && changed > 0 {
		return exitError
	}
	return exitOK
}

// runStdout strips one file and writes the result to stdout.
func runStdout(path string, cfg config, opt scan.Options) int {
	// -stdout does not go through walk, so the credential guard has to be
	// repeated here. Printing a private key is not a write, but it still
	// puts the contents somewhere the user did not ask for.
	if !cfg.allowSecrets {
		if why := walk.SecretReason(path); why != "" {
			fmt.Fprintf(os.Stderr, "%s: %s (use -allow-secrets)\n", path, why)
			return exitError
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	if textio.IsBinary(raw) {
		fmt.Fprintf(os.Stderr, "%s: binary file\n", path)
		return exitError
	}
	l := cfg.lang
	if l == nil {
		l = lang.ByFilename(filepath.Base(path))
	}
	if l == nil {
		fmt.Fprintf(os.Stderr, "%s: unknown language, use -lang\n", path)
		return exitError
	}
	text, format := textio.Decode(raw)
	sr := scan.Strip(text, l, opt)
	if sr.Unsafe && !cfg.force {
		fmt.Fprintf(os.Stderr, "%s: %s (use -force)\n", path, sr.Reason)
		return exitError
	}
	if !cfg.noVerify {
		if err := verify.Check(text, sr.Output, l); err != nil {
			fmt.Fprintf(os.Stderr, "%s: verification failed: %v\n", path, err)
			return exitError
		}
	}
	if _, err := os.Stdout.Write(textio.Encode(sr.Output, format)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	return exitOK
}

// printLanguages lists every supported language.
func printLanguages() {
	for _, l := range lang.All() {
		risky := ""
		if l.Risky {
			risky = "  (risky)"
		}
		names := append(append([]string{}, l.Exts...), l.Filenames...)
		fmt.Printf("%-12s %s%s\n", l.Name, strings.Join(names, " "), risky)
	}
}

// displayPath prefers a path relative to the working directory.
func displayPath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

// splitList parses a comma separated flag value.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
