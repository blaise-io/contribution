package main

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Graph is a collection of pixels to draw.
type Graph struct {
	bounds image.Rectangle
	pixels []GraphPixel
}

// GraphPixel is a single pixel to draw.
type GraphPixel struct {
	daysAgo  int
	darkness int
}

// PixelChr returns unicode characters resembling a single GitHub graph pixel.
func PixelChr(darkness int) string {
	r := []rune(" ░▒▓█")
	return strings.Repeat(string(r[darkness]), 2) + " "
}

// ImageToGraphPixels converts an image to an array of GraphPixels.
func ImageToGraphPixels(colorMapped ColorMappedImage, args Args) Graph {
	bounds := colorMapped.image.Bounds()
	graph := Graph{bounds: bounds}
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			gray := colorMapped.image.GrayAt(x, y)
			commits := colorMapped.colorMap[int(gray.Y)]
			graph.pixels = append(graph.pixels, GraphPixel{
				pixelDaysAgo(x, y, bounds.Max.X, args.weeksAgo), commits,
			})
		}
	}
	return graph
}

func getBaseSSHCmd() []string {
	envData := os.Getenv("GIT_SSH_COMMAND")
	if envData != "" {
		return strings.Split(envData, " ")
	}
	return []string{"ssh"}
}

// GetUsername gets the GitHub username of the user
func GetUsername() (string, error) {
	cmd := append(getBaseSSHCmd(), "-T", "git@github.com")
	resp, _ := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()

	scan := strings.Split(string(resp), "!")[0]
	var username string
	_, err := fmt.Sscanf(scan, "Hi %s", &username)
	if err != nil {
		return "", errors.New("Invalid identity")
	}
	return username, nil
}

// CreateBranch creates a git branch locally.
func CreateBranch(dir string, branch string) error {
	return commands(
		dir,
		[]string{"git", "init"},
		[]string{"git", "checkout", "-B", branch},
	)
}

// CommitAll converts and executes all GraphPixels to git commits.
func CommitAll(dir string, graph Graph, args Args) error {
	chrs := ""
	filename := filepath.Join(dir, "README.md")
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Println()

	for i, px := range graph.pixels {
		err = drawPixel(dir, f, px, chrs)
		if err != nil {
			return err
		}
		fmt.Print(PixelChr(px.darkness))
		chrs += PixelChr(px.darkness)
		if (i+1)%graph.bounds.Max.X == 0 {
			chrs += "\n"
			fmt.Println()
		}
	}
	fmt.Println()
	return nil
}

// PushAll pushes all local commits to remote git.
func PushAll(dir string, project string, branch string) error {
	githubURL := fmt.Sprintf("git@github.com:%s.git", project)
	return command(dir, "git", "push", "-fu", githubURL, branch)
}

func pixelDaysAgo(x int, y int, width int, weeksAgo int) int {
	startDaysAgo := (width + weeksAgo) * 7
	startDaysAgo += int(time.Now().Weekday())
	return startDaysAgo - (x * 7) - y
}

func drawPixel(dir string, f *os.File, px GraphPixel, pixelChrs string) error {
	var err error
	multiply := 2

	for i := 1; i <= px.darkness; i++ {
		// Multiply the number of commits per pixel to reduce the noise
		// of any other user activity in the user's activity graph.
		for j := multiply; j > 0; j-- {
			s := markdownStr(pixelChrs+PixelChr(i), j)

			// Write changes to file so it can be git committed.
			f.Seek(0, io.SeekStart)
			_, err = f.Write([]byte(s))
			if err != nil {
				return err
			}

			// Commit changes.
			err = commit(dir, px.daysAgo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// markDownstr wraps the block char pixels in a code section for
// monospace display and adds a temporary `<` suffix to achieve unique
// file contents to write for multiple commits per pixel.
func markdownStr(pixelChrs string, ncommits int) string {
	s := "```\n" + pixelChrs
	if ncommits > 1 {
		s += strings.Repeat("<", ncommits-1)
	}
	return s + "\n```\n"
}

func commit(dir string, daysAgo int) error {
	date := time.Now().AddDate(0, 0, -daysAgo).Format(time.RFC3339)
	return commands(dir,
		[]string{"git", "add", "."},
		[]string{
			"git", "commit",
			"--all", "--allow-empty-message",
			"--message", "",
			"--date", date,
		},
	)
}

// command executes a single command in a custom environment and dir.
func command(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	response, err := cmd.CombinedOutput()
	responseStr := string(response)
	if err != nil {
		return fmt.Errorf("%s %v: %v", name, arg, responseStr)
	}
	return nil
}

// commands executes multiple commands.
func commands(dir string, cmds ...[]string) error {
	for _, cmd := range cmds {
		err := command(dir, cmd[0], cmd[1:]...)
		if err != nil {
			return err
		}
	}
	return nil
}
