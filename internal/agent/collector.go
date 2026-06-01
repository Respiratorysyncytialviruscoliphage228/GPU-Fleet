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

var richGPUQueryFields = []string{
	"index",
	"name",
	"uuid",
	"driver_version",
	"vbios_version",
	"memory.total",
	"memory.used",
	"memory.free",
	"memory.reserved",
	"utilization.gpu",
	"utilization.memory",
	"temperature.gpu",
	"temperature.memory",
	"temperature.gpu.tlimit",
	"power.draw",
	"power.limit",
	"enforced.power.limit",
	"fan.speed",
	"clocks.gr",
	"clocks.mem",
	"clocks.sm",
	"clocks.video",
	"pstate",
	"pcie.link.gen.current",
	"pcie.link.gen.max",
	"pcie.link.width.current",
	"pcie.link.width.max",
	"compute_mode",
	"compute_cap",
	"display_active",
	"display_attached",
	"persistence_mode",
	"driver_model.current",
	"ecc.mode.current",
	"mig.mode.current",
	"clocks_event_reasons.active",
}

var basicGPUQueryFields = []string{
	"index",
	"name",
	"uuid",
	"driver_version",
	"memory.total",
	"memory.used",
	"utilization.gpu",
	"temperature.gpu",
	"power.draw",
	"fan.speed",
	"clocks.gr",
	"clocks.mem",
	"pstate",
	"pcie.link.gen.current",
	"pcie.link.width.current",
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

	fields := richGPUQueryFields
	output, err := c.queryGPU(ctx, fields)
	if err != nil {
		fields = basicGPUQueryFields
		output, err = c.queryGPU(ctx, fields)
	}
	if err != nil {
		return model.GPUSample{
			Timestamp: time.Now().UTC(),
			GPUs: []model.GPUStatus{{
				GPUID:           "gpu0",
				CollectionError: collectionError(err, ctx.Err()),
			}},
		}, err
	}

	return parseGPUSample(output, fields)
}

func (c Collector) queryGPU(ctx context.Context, fields []string) ([]byte, error) {
	args := []string{
		"--query-gpu=" + strings.Join(fields, ","),
		"--format=csv,noheader,nounits",
	}
	cmd := exec.CommandContext(ctx, c.Command, args...)
	return cmd.Output()
}

func parseGPUSample(output []byte, fields []string) (model.GPUSample, error) {
	reader := csv.NewReader(bytes.NewReader(output))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return model.GPUSample{}, err
	}

	sample := model.GPUSample{Timestamp: time.Now().UTC()}
	for _, record := range records {
		if len(record) < len(fields) {
			continue
		}
		values := map[string]string{}
		for index, field := range fields {
			values[field] = record[index]
		}
		totalMiB := parseUint(values["memory.total"])
		usedMiB := parseUint(values["memory.used"])
		freeMiB := parseUint(values["memory.free"])
		reservedMiB := parseUint(values["memory.reserved"])
		status := model.GPUStatus{
			GPUID:                 "gpu" + clean(values["index"]),
			Name:                  clean(values["name"]),
			UUIDHash:              "sha256:" + auth.SHA256Hex(clean(values["uuid"])),
			DriverVersion:         clean(values["driver_version"]),
			VBIOSVersion:          clean(values["vbios_version"]),
			MemoryTotalBytes:      totalMiB * 1024 * 1024,
			MemoryUsedBytes:       usedMiB * 1024 * 1024,
			MemoryFreeBytes:       freeMiB * 1024 * 1024,
			MemoryReservedBytes:   reservedMiB * 1024 * 1024,
			PCIeLinkGeneration:    clean(values["pcie.link.gen.current"]),
			PCIeLinkGenerationMax: clean(values["pcie.link.gen.max"]),
			PCIeLinkWidth:         clean(values["pcie.link.width.current"]),
			PCIeLinkWidthMax:      clean(values["pcie.link.width.max"]),
			PState:                clean(values["pstate"]),
			ComputeMode:           clean(values["compute_mode"]),
			ComputeCapability:     clean(values["compute_cap"]),
			DisplayActive:         clean(values["display_active"]),
			DisplayAttached:       clean(values["display_attached"]),
			PersistenceMode:       clean(values["persistence_mode"]),
			DriverModel:           clean(values["driver_model.current"]),
			ECCModeCurrent:        clean(values["ecc.mode.current"]),
			MIGModeCurrent:        clean(values["mig.mode.current"]),
			ClockThrottleReasons:  clean(values["clocks_event_reasons.active"]),
		}
		status.UtilizationGPUPercent = parseFloatPtr(values["utilization.gpu"])
		status.UtilizationMemPercent = parseFloatPtr(values["utilization.memory"])
		status.TemperatureCelsius = parseFloatPtr(values["temperature.gpu"])
		status.TemperatureMemCelsius = parseFloatPtr(values["temperature.memory"])
		status.TemperatureLimitC = parseFloatPtr(values["temperature.gpu.tlimit"])
		status.PowerDrawWatts = parseFloatPtr(values["power.draw"])
		status.PowerLimitWatts = parseFloatPtr(values["power.limit"])
		status.PowerEnforcedLimitW = parseFloatPtr(values["enforced.power.limit"])
		status.FanSpeedPercent = parseFloatPtr(values["fan.speed"])
		status.GraphicsClockMHz = parseFloatPtr(values["clocks.gr"])
		status.MemoryClockMHz = parseFloatPtr(values["clocks.mem"])
		status.SMClockMHz = parseFloatPtr(values["clocks.sm"])
		status.VideoClockMHz = parseFloatPtr(values["clocks.video"])
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
