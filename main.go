package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
)

func preview(img image.Image) error {
	PreviewResult(img)
	SavePNG(ToGithubPalette(img), "./example/preview.png")
	return nil
}

func push(img image.Image, args Args) error {
	username, err := GetUsername()
	if err != nil {
		return err
	}

	colorMapped := getColorMappedImage(img)
	commits := ImageToGraphPixels(colorMapped, args)

	dir, err := ioutil.TempDir("", filepath.Base(os.Args[0]))
	if err != nil {
		return err
	}

	err = CreateBranch(dir, args.githubBranch)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	err = CommitAll(dir, commits, args)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Pushing... ")
	err = PushAll(dir, args.githubProject, args.githubBranch)
	if err != nil {
		return err
	}

	fmt.Println("Done!")
	fmt.Printf("Go check out https://github.com/%s\n", username)
	return nil
}

func main() {
	args := GetArgs()

	img, err := ReadImage(args.imageFile)
	if err != nil {
		exitWithError(err)
	}
	img = ResizeImage(img)

	if args.verb == "preview" {
		err = preview(img)
	} else if args.verb == "push" {
		err = push(img, args)
	}
	if err != nil {
		exitWithError(err)
	}

	os.Exit(0)
}

func exitWithError(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}
