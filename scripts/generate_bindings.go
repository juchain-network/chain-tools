package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type artifact struct {
	ABI      any `json:"abi"`
	Bytecode struct {
		Object string `json:"object"`
	} `json:"bytecode"`
}

func main() {
	var (
		artifactPath = flag.String("artifact", "", "Path to Foundry contract artifact JSON")
		packageName  = flag.String("pkg", "contracts", "Go package name")
		typeName     = flag.String("type", "", "Go binding type name")
		outputPath   = flag.String("out", "", "Output Go file path")
	)
	flag.Parse()

	if *artifactPath == "" {
		exitf("--artifact is required")
	}
	if *typeName == "" {
		exitf("--type is required")
	}
	if *outputPath == "" {
		exitf("--out is required")
	}

	data, err := os.ReadFile(*artifactPath)
	if err != nil {
		exitf("read artifact %s: %v", *artifactPath, err)
	}

	var parsed artifact
	if err := json.Unmarshal(data, &parsed); err != nil {
		exitf("unmarshal artifact %s: %v", *artifactPath, err)
	}
	if parsed.ABI == nil {
		exitf("artifact %s does not contain abi", *artifactPath)
	}
	if parsed.Bytecode.Object == "" {
		exitf("artifact %s does not contain bytecode.object", *artifactPath)
	}

	abiJSON, err := json.Marshal(parsed.ABI)
	if err != nil {
		exitf("marshal abi %s: %v", *artifactPath, err)
	}

	code, err := bind.Bind(
		[]string{*typeName},
		[]string{string(abiJSON)},
		[]string{parsed.Bytecode.Object},
		nil,
		*packageName,
		nil,
		nil,
	)
	if err != nil {
		exitf("generate binding for %s: %v", *artifactPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		exitf("create output dir for %s: %v", *outputPath, err)
	}
	if err := os.WriteFile(*outputPath, []byte(code), 0o644); err != nil {
		exitf("write output %s: %v", *outputPath, err)
	}
}

func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
