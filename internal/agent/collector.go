package agent

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gpufleet/internal/auth"
	"gpufleet/internal/model"
)

type Collector struct {
	Command string
	Timeout time.Duration
}

func NewCollector(command string, timeout time.Duration) Collector {
	if command == "" {
		command = "nvidia-smi"
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return Collector{Command: command, Timeout: timeout}
}

func (c Collector) Collect(ctx context.Context) (model.GPUSample, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	args := []string{
		"--query-gpu=index,name,uuid,driver_version,memory.total,memory.used,utilization.gpu,temperature.gpu,power.draw,fan.speed,clocks.gr,clocks.mem,pstate,pcie.link.gen.current,pcie.link.width.current",
		"--format=csv,noheader,nounits",
	}
	cmd := exec.CommandContext(ctx, c.Command, args...)
	output, err := cmd.Output()
	if err != nil {
		return model.GPUSample{
			Timestamp: time.Now().UTC(),
			GPUs: []model.GPUStatus{{
				GPUID:           "gpu0",
				CollectionError: collectionError(err, ctx.Err()),
			}},
		}, err
	}

	reader := csv.NewReader(bytes.NewReader(output))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return model.GPUSample{}, err
	}

	sample := model.GPUSample{Timestamp: time.Now().UTC()}
	for _, record := range records {
		if len(record) < 15 {
			continue
		}
		totalMiB := parseUint(record[4])
		usedMiB := parseUint(record[5])
		status := model.GPUStatus{
			GPUID:              "gpu" + clean(record[0]),
			Name:               clean(record[1]),
			UUIDHash:           "sha256:" + auth.SHA256Hex(clean(record[2])),
			DriverVersion:      clean(record[3]),
			MemoryTotalBytes:   totalMiB * 1024 * 1024,
			MemoryUsedBytes:    usedMiB * 1024 * 1024,
			PCIeLinkGeneration: clean(record[13]),
			PCIeLinkWidth:      clean(record[14]),
			PState:             clean(record[12]),
		}
		status.UtilizationGPUPercent = parseFloatPtr(record[6])
		status.TemperatureCelsius = parseFloatPtr(record[7])
		status.PowerDrawWatts = parseFloatPtr(record[8])
		status.FanSpeedPercent = parseFloatPtr(record[9])
		status.GraphicsClockMHz = parseFloatPtr(record[10])
		status.MemoryClockMHz = parseFloatPtr(record[11])
		sample.GPUs = append(sample.GPUs, status)
	}
	return sample, nil
}

func (c Collector) CollectProcesses(ctx context.Context, sample model.GPUSample) (model.ProcessBatch, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	uuidToGPU := map[string]string{}
	for _, gpu := range sample.GPUs {
		if gpu.UUIDHash != "" && gpu.GPUID != "" {
			uuidToGPU[gpu.UUIDHash] = gpu.GPUID
		}
	}

	args := []string{
		"--query-compute-apps=gpu_uuid,pid,process_name,used_memory",
		"--format=csv,noheader,nounits",
	}
	cmd := exec.CommandContext(ctx, c.Command, args...)
	output, err := cmd.Output()
	if err != nil {
		return model.ProcessBatch{Timestamp: time.Now().UTC()}, err
	}

	reader := csv.NewReader(bytes.NewReader(output))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return model.ProcessBatch{}, err
	}

	batch := model.ProcessBatch{Timestamp: time.Now().UTC()}
	for _, record := range records {
		if len(record) < 4 {
			continue
		}
		uuidHash := "sha256:" + auth.SHA256Hex(clean(record[0]))
		processName := clean(record[2])
		if processName == "" {
			processName = "unknown"
		} else {
			processName = filepath.Base(processName)
		}
		snapshot := model.ProcessSnapshot{
			GPUID:           uuidToGPU[uuidHash],
			UUIDHash:        uuidHash,
			PID:             int(parseUint(record[1])),
			ProcessName:     processName,
			UsedMemoryBytes: parseUint(record[3]) * 1024 * 1024,
		}
		if snapshot.PID == 0 {
			continue
		}
		batch.Processes = append(batch.Processes, snapshot)
	}
	return batch, nil
}

func clean(value string) string {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "[not supported]") ||
		strings.EqualFold(value, "not supported") ||
		strings.EqualFold(value, "N/A") ||
		strings.EqualFold(value, "[N/A]") {
		return ""
	}
	return value
}

func parseFloatPtr(value string) *float64 {
	cleaned := clean(value)
	if cleaned == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseUint(value string) uint64 {
	cleaned := clean(value)
	if cleaned == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(cleaned, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func collectionError(err error, ctxErr error) string {
	if ctxErr != nil {
		return "collection_timeout"
	}
	if err == nil {
		return ""
	}
	return fmt.Sprintf("nvidia_smi_failed: %v", err)
}
