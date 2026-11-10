// Command dump prints the full annotation of the verses matching a phrase.
// It is a development aid for reviewing the annotation quality.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sumate/sumate/internal/annotate"
	"github.com/sumate/sumate/internal/corpus"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: dump <phrase>")
		os.Exit(1)
	}
	c, err := corpus.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	verses, err := corpus.NewPipeline(c, annotate.New()).Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	n := 0
	for _, v := range verses {
		if !strings.Contains(v.Text, os.Args[1]) {
			continue
		}
		n++
		if n > 2 {
			break
		}
		fmt.Printf("=== %s (%s) ===\n", v.Text, v.Tune)
		fmt.Printf("题:   %s\n", v.Title)
		fmt.Printf("背景: %s | %s | %s\n", v.Background.Year, v.Background.Place, v.Background.Historical)
		fmt.Printf("表层: %s\n", v.Annotation.Literal)
		fmt.Printf("深意: %s\n", v.Annotation.Conception)
		fmt.Printf("哲思: %s\n", v.Annotation.Philosophy)
		fmt.Printf("态度: %s\n", v.Annotation.Attitude)
		fmt.Printf("极性: %s\n", v.Annotation.Polarity)
		fmt.Printf("人群: %s\n", v.Annotation.Crowd)
		fmt.Printf("阈值: %s | 边界: %s\n", v.Annotation.MoodThreshold, v.Annotation.ScenarioBoundary)
		fmt.Printf("疗愈: %s\n", v.Annotation.Healing)
		fmt.Printf("概念: %v\n", v.Annotation.Concepts)
		fmt.Printf("情绪: ")
		for _, e := range v.Annotation.Emotions {
			fmt.Printf("[%s/%s/%s] ", e.Dimension, e.Name, e.Strength)
		}
		fmt.Printf("\n气质: %v\n", v.Annotation.Tangents)
		fmt.Printf("标签: %v\n\n", v.Annotation.Tags)
	}
}
