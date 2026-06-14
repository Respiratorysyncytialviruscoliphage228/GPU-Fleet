package agent

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gpufleet/internal/model"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientReturnsHTTPStatusError(t *testing.T) {
	client := &Client{
		ServerURL: "http://example.test",
		DeviceID:  "device-test",
		Secret:    "secret-test",
		HTTP: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			_, _ = io.Copy(io.Discard, req.Body)
			_ = req.Body.Close()
			return httpStatusResponse(req, http.StatusTooManyRequests, `{"error":"too many agent requests"}`), nil
		})},
	}

	err := client.PostHeartbeat(model.Heartbeat{Timestamp: time.Now().UTC()})
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected HTTPStatusError, got %T %v", err, err)
	}
	if statusErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", statusErr.StatusCode)
	}
	if !strings.Contains(statusErr.Error(), "too many agent requests") {
		t.Fatalf("expected response body in error, got %q", statusErr.Error())
	}
}

func TestSampleQueueFlushReturnsRetryableUploadError(t *testing.T) {
	queue, err := NewSampleQueue(t.TempDir(), 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(testBatchAt(time.Unix(100, 0))); err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(testBatchAt(time.Unix(110, 0))); err != nil {
		t.Fatal(err)
	}

	calls := 0
	client := testClient(func(req *http.Request) (*http.Response, error) {
		calls++
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
		return httpStatusResponse(req, http.StatusServiceUnavailable, "offline"), nil
	})

	err = queue.Flush(client, 10)
	if err == nil {
		t.Fatal("expected retryable upload error")
	}
	if calls != 1 {
		t.Fatalf("expected one upload attempt, got %d", calls)
	}
	if got := readQueuedBatches(t, queue); len(got) != 2 {
		t.Fatalf("expected both batches to remain queued, got %d", len(got))
	}
}

func TestSampleQueueFlushDropsPermanentPayloadError(t *testing.T) {
	queue, err := NewSampleQueue(t.TempDir(), 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(testBatchAt(time.Unix(100, 0))); err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(testBatchAt(time.Unix(110, 0))); err != nil {
		t.Fatal(err)
	}

	calls := 0
	client := testClient(func(req *http.Request) (*http.Response, error) {
		calls++
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
		if calls == 1 {
			return httpStatusResponse(req, http.StatusBadRequest, `{"error":"device id mismatch"}`), nil
		}
		return httpStatusResponse(req, http.StatusAccepted, `{"accepted":true}`), nil
	})

	if err := queue.Flush(client, 10); err != nil {
		t.Fatalf("expected permanent bad batch to be dropped, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected two upload attempts, got %d", calls)
	}
	if got := readQueuedBatches(t, queue); len(got) != 0 {
		t.Fatalf("expected queue to drain after dropping bad batch, got %d", len(got))
	}
}

func testClient(roundTrip roundTripFunc) *Client {
	return &Client{
		ServerURL: "http://example.test",
		DeviceID:  "device-test",
		Secret:    "secret-test",
		HTTP:      &http.Client{Transport: roundTrip},
	}
}

func httpStatusResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func testBatchAt(at time.Time) model.SampleBatch {
	return model.SampleBatch{
		DeviceID:     "device-test",
		AgentVersion: "test",
		Samples: []model.GPUSample{{
			Timestamp: at.UTC(),
			GPUs: []model.GPUStatus{{
				GPUID: "gpu0",
				Name:  "test-gpu",
			}},
		}},
	}
}

func readQueuedBatches(t *testing.T, queue *SampleQueue) []model.SampleBatch {
	t.Helper()
	queue.mu.Lock()
	defer queue.mu.Unlock()
	batches, err := queue.readLocked()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return batches
}
