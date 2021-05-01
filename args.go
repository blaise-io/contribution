package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Args contains represents CLI arguments for this program.
type Args struct {
	verb          string
	imageFile     string
	githubBranch  string
	githubProject string
	weeksAgo      int
}

// GetArgs returns CLI arguments.
func GetArgs() Args {
	args := Args{}

	preview := flag.NewFlagSet("preview", flag.ExitOnError)
	preview.Usage = func() {
		exitWithPreviewHelp(preview)
	}

	push := flag.NewFlagSet("push", flag.ExitOnError)
	push.StringVar(&args.githubBranch, "branch", "contribution", "Git branch to push to.")
	push.IntVar(&args.weeksAgo, "w", 0, "Weeks ago of all activity. A value of 2 will move activity two pixels to the left.")
	push.StringVar(&args.githubProject, "project", "", "GitHub username/project to push to. (required)")
	push.Usage = func() {
		exitWithPushHelp(push)
	}

	for _, flagset := range []*flag.FlagSet{preview, push} {
		flagset.StringVar(&args.imageFile, "img", "", "Path to a valid PNG image. (required)")
	}

	subcmd := ""
	if len(os.Args) >= 2 {
		subcmd = os.Args[1]
	}

	switch subcmd {
	case preview.Name():
		preview.Parse(os.Args[2:])
		args.verb = preview.Name()
	case push.Name():
		push.Parse(os.Args[2:])
		args.verb = push.Name()
		if args.githubProject == "" {
			exitWithPushHelp(push)
		}
	default:
		exitWithHelp(preview, push)
	}

	return args
}

// exitWithHelp exits with global help
func exitWithHelp(preview *flag.FlagSet, push *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Draw an image on your GitHub contribution history.\n")
	fmt.Fprintf(os.Stderr, "By Blaise Kal, %d\n\n", time.Now().Year())
	fmt.Fprintf(os.Stderr, "Preview contribution history graph without pushing\n")
	fmt.Fprintf(os.Stderr, "  %s %s -img /path/to/image.png\n", filepath.Base(os.Args[0]), preview.Name())
	fmt.Fprintf(os.Stderr, "Preview usage and options\n")
	fmt.Fprintf(os.Stderr, "  %s %s -help\n\n", filepath.Base(os.Args[0]), preview.Name())
	fmt.Fprintf(os.Stderr, "Push contribution history graph to GitHub\n")
	fmt.Fprintf(os.Stderr, "  %s %s -img /path/to/image.png -project username/project\n", filepath.Base(os.Args[0]), push.Name())
	fmt.Fprintf(os.Stderr, "Push usage and options\n")
	fmt.Fprintf(os.Stderr, "  %s %s -help\n", filepath.Base(os.Args[0]), push.Name())
	os.Exit(0)
}

// helpPreview exits with help on the preview subcommand
func exitWithPreviewHelp(flagset *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Preview contribution history graph without pushing.\n")
	fmt.Fprintf(os.Stderr, "Example usage:\n")
	fmt.Fprintf(os.Stderr, "  %s %s -img image.png\n\n", filepath.Base(os.Args[0]), os.Args[1])
	flag.PrintDefaults()
	flagset.PrintDefaults()
	os.Exit(0)
}

// helpPreview exits with help on the push subcommand
func exitWithPushHelp(flagset *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Push contribution history graph to GitHub\n")
	fmt.Fprintf(os.Stderr, "Example usage:\n")
	fmt.Fprintf(os.Stderr, "  %s %s -project username/project\n\n", filepath.Base(os.Args[0]), os.Args[1])
	flag.PrintDefaults()
	flagset.PrintDefaults()
	os.Exit(0)
}
