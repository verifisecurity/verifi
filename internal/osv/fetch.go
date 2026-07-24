package osv

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// osvBucket is the public OSV data bucket. Each ecosystem publishes an all.zip
// of its advisories. See https://google.github.io/osv.dev/data/#data-dumps.
const osvBucket = "https://osv-vulnerabilities.storage.googleapis.com"

// Fetch downloads the OSV advisory database for an ecosystem and extracts the
// JSON advisories into cacheRoot/<ecosystem>/, replacing what was there. It
// returns the number of advisories written. After this, matching is offline.
func Fetch(ecosystem, cacheRoot string) (int, error) {
	url := fmt.Sprintf("%s/%s/all.zip", osvBucket, ecosystem)
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	// Stream to a temp file so a large dump does not sit in memory; zip needs
	// random access, which a file provides.
	tmp, err := os.CreateTemp("", "verifi-osv-*.zip")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return 0, fmt.Errorf("save download: %w", err)
	}
	fi, err := tmp.Stat()
	if err != nil {
		return 0, err
	}
	return extractZip(tmp, fi.Size(), filepath.Join(cacheRoot, ecosystem))
}

// extractZip writes every .json entry in the zip to destDir, flattened to base
// names, replacing whatever was there. Separated from Fetch so it is testable
// without the network.
func extractZip(r io.ReaderAt, size int64, destDir string) (int, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return 0, fmt.Errorf("open zip: %w", err)
	}
	if err := os.RemoveAll(destDir); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}

	n := 0
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(f.Name, ".json") {
			continue
		}
		if err := writeEntry(f, filepath.Join(destDir, filepath.Base(f.Name))); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func writeEntry(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, rc); err != nil {
		return err
	}
	return nil
}
