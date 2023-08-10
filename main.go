package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/itchyny/gojq"
	"github.com/yseto/check-json/reader"
	"github.com/yseto/check-json/state"
)

var sig = os.Interrupt

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), sig)
	defer stop()

	var inputQuery = flag.String("query", "", "query of jq style.")
	var filename = flag.String("log-file", "", "logfile")
	var stateDir = flag.String("state-dir", os.TempDir(), "state file directory")
	var noState = flag.Bool("no-state", false, "do not use state file")
	var outputFormat = flag.String("output", ".", "output format of jq style")

	flag.Parse()

	if *inputQuery == "" || *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	query, err := gojq.Parse(*inputQuery)
	if err != nil {
		log.Fatalln(err)
	}

	var stateHandle reader.StateHolder
	if *noState {
		stateHandle = state.Empty()
	} else {
		stateHandle = state.New(filepath.Join(*stateDir, stateFilename(*filename)))
	}

	err = reader.New(*filename, stateHandle).Read(ctx, func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				obj := make(map[string]any, 0)
				json.Unmarshal(scanner.Bytes(), &obj)
				eval(ctx, query, obj, *outputFormat)
			}
		}
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func eval(ctx context.Context, query *gojq.Query, input any, format string) {
	iter := query.RunWithContext(ctx, input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if _, ok := v.(error); ok {
			// log.Println(err)
			continue
		}
		// fmt.Printf("%#v\n", v)
		output(ctx, format, v)
		// fmt.Println("")
	}
}

func output(ctx context.Context, format string, input any) {
	query, err := gojq.Parse(format)
	if err != nil {
		log.Fatalln(err)
	}
	iter := query.RunWithContext(ctx, input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if _, ok := v.(error); ok {
			continue
		}

		b, _ := json.Marshal(v)
		fmt.Println(string(b))
	}
}

func stateFilename(filename string) string {
	h := sha1.New()
	h.Write([]byte(filename))
	return fmt.Sprintf("check-json-%x.json", h.Sum(nil))
}
