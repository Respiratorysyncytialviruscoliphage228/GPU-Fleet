package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gpufleet/internal/agent"
)

func main() {
	var version string
	var assetsDir string
	var releaseBaseURL string
	var privateKey string
	var outputPath string
	var publicKeyOutputPath string
	var minAgentVersion string

	flag.StringVar(&version, "version", "", "release version without the v prefix")
	flag.StringVar(&assetsDir, "assets-dir", "dist/release", "directory containing raw Agent release assets")
	flag.StringVar(&releaseBaseURL, "release-base-url", "", "base URL for release assets")
	flag.StringVar(&privateKey, "private-key", os.Getenv("GPUFLEET_AGENT_UPDATE_ED25519_PRIVATE_KEY"), "base64 Ed25519 private key seed or private key")
	flag.StringVar(&outputPath, "output", "dist/release/gpufleet-agent-manifest.json", "output manifest path")
	flag.StringVar(&publicKeyOutputPath, "public-key-output", "", "optional output path for the base64 Ed25519 public key")
	flag.StringVar(&minAgentVersion, "min-agent-version", "", "optional minimum Agent version")
	flag.Parse()

	if strings.TrimSpace(version) == "" {
		fatalf("version is required")
	}
	if strings.TrimSpace(releaseBaseURL) == "" {
		fatalf("release-base-url is required")
	}
	private, public := parsePrivateKey(privateKey)
	artifacts, err := scanArtifacts(assetsDir, version, strings.TrimRight(releaseBaseURL, "/"))
	if err != nil {
		fatalf("%v", err)
	}
	if len(artifacts) == 0 {
		fatalf("no raw Agent update assets found in %s for version %s", assetsDir, version)
	}

	manifest := agent.UpdateManifest{
		Version:         version,
		CreatedAt:       time.Now().UTC(),
		MinAgentVersion: strings.TrimSpace(minAgentVersion),
		Artifacts:       artifacts,
	}
	payload, err := manifest.SigningBytes()
	if err != nil {
		fatalf("canonicalize manifest: %v", err)
	}
	manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload))
	if err := manifest.Verify(base64.StdEncoding.EncodeToString(public)); err != nil {
		fatalf("signed manifest did not verify: %v", err)
	}

	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fatalf("encode manifest: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fatalf("create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, append(raw, '\n'), 0644); err != nil {
		fatalf("write manifest: %v", err)
	}
	if publicKeyOutputPath != "" {
		if err := os.MkdirAll(filepath.Dir(publicKeyOutputPath), 0755); err != nil {
			fatalf("create public key output directory: %v", err)
		}
		if err := os.WriteFile(publicKeyOutputPath, []byte(base64.StdEncoding.EncodeToString(public)+"\n"), 0644); err != nil {
			fatalf("write public key: %v", err)
		}
	}
	fmt.Printf("wrote %s with %d artifacts\n", outputPath, len(artifacts))
}

func parsePrivateKey(raw string) (ed25519.PrivateKey, ed25519.PublicKey) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		fatalf("GPUFLEET_AGENT_UPDATE_ED25519_PRIVATE_KEY is required")
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		fatalf("decode private key: %v", err)
	}
	var private ed25519.PrivateKey
	switch len(decoded) {
	case ed25519.SeedSize:
		private = ed25519.NewKeyFromSeed(decoded)
	case ed25519.PrivateKeySize:
		private = ed25519.PrivateKey(decoded)
	default:
		fatalf("private key must be %d-byte seed or %d-byte private key", ed25519.SeedSize, ed25519.PrivateKeySize)
	}
	public, ok := private.Public().(ed25519.PublicKey)
	if !ok || len(public) != ed25519.PublicKeySize {
		fatalf("derive public key failed")
	}
	return private, public
}

func scanArtifacts(dir, version, releaseBaseURL string) ([]agent.UpdateArtifact, error) {
	pattern := regexp.MustCompile(`^gpufleet-agent_` + regexp.QuoteMeta(version) + `_([^_]+)_([^._]+)(?:\.exe)?$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var artifacts []agent.UpdateArtifact
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		match := pattern.FindStringSubmatch(name)
		if match == nil {
			continue
		}
		arch, ok := manifestArch(match[2])
		if !ok {
			continue
		}
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(raw)
		artifacts = append(artifacts, agent.UpdateArtifact{
			OS:        match[1],
			Arch:      arch,
			URL:       releaseBaseURL + "/" + name,
			SHA256:    hex.EncodeToString(sum[:]),
			SizeBytes: int64(len(raw)),
			Filename:  name,
		})
	}
	sort.Slice(artifacts, func(i, j int) bool {
		if artifacts[i].OS == artifacts[j].OS {
			return artifacts[i].Arch < artifacts[j].Arch
		}
		return artifacts[i].OS < artifacts[j].OS
	})
	return artifacts, nil
}

func manifestArch(label string) (string, bool) {
	if strings.HasPrefix(label, "armv") {
		return "", false
	}
	return label, true
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
