package agent

import "testing"

func TestParseGPUSampleRichFields(t *testing.T) {
	output := []byte("0, NVIDIA GeForce RTX 5060 Ti, GPU-test, 591.74, 98.06.1f.00.c8, 16311, 5260, 10791, 261, 100, 70, 72, N/A, 15, 180.29, 180.00, 180.00, 84, 2602, 16001, 2602, 2265, P0, 3, 3, 8, 16, Default, 12.0, Disabled, No, [N/A], WDDM, [N/A], [N/A], 0x0000000000000000\n")

	sample, err := parseGPUSample(output, richGPUQueryFields)
	if err != nil {
		t.Fatal(err)
	}
	if len(sample.GPUs) != 1 {
		t.Fatalf("expected one GPU, got %d", len(sample.GPUs))
	}
	gpu := sample.GPUs[0]
	if gpu.GPUID != "gpu0" {
		t.Fatalf("expected gpu0, got %q", gpu.GPUID)
	}
	if gpu.Name != "NVIDIA GeForce RTX 5060 Ti" {
		t.Fatalf("unexpected name %q", gpu.Name)
	}
	if gpu.VBIOSVersion != "98.06.1f.00.c8" {
		t.Fatalf("unexpected vbios %q", gpu.VBIOSVersion)
	}
	if gpu.MemoryFreeBytes != 10791*1024*1024 {
		t.Fatalf("unexpected free memory %d", gpu.MemoryFreeBytes)
	}
	if gpu.UtilizationMemPercent == nil || *gpu.UtilizationMemPercent != 70 {
		t.Fatalf("unexpected memory utilization %+v", gpu.UtilizationMemPercent)
	}
	if gpu.TemperatureMemCelsius != nil {
		t.Fatalf("expected unsupported memory temperature to be nil, got %+v", gpu.TemperatureMemCelsius)
	}
	if gpu.PowerLimitWatts == nil || *gpu.PowerLimitWatts != 180 {
		t.Fatalf("unexpected power limit %+v", gpu.PowerLimitWatts)
	}
	if gpu.SMClockMHz == nil || *gpu.SMClockMHz != 2602 {
		t.Fatalf("unexpected SM clock %+v", gpu.SMClockMHz)
	}
	if gpu.PCIeLinkWidthMax != "16" {
		t.Fatalf("unexpected max PCIe width %q", gpu.PCIeLinkWidthMax)
	}
	if gpu.ComputeCapability != "12.0" {
		t.Fatalf("unexpected compute capability %q", gpu.ComputeCapability)
	}
	if gpu.DriverModel != "WDDM" {
		t.Fatalf("unexpected driver model %q", gpu.DriverModel)
	}
	if gpu.ECCModeCurrent != "" || gpu.MIGModeCurrent != "" {
		t.Fatalf("expected unsupported ECC/MIG fields to be empty, got %q/%q", gpu.ECCModeCurrent, gpu.MIGModeCurrent)
	}
}

func TestParseGPUSampleBasicFields(t *testing.T) {
	output := []byte("0, NVIDIA GeForce RTX 5060 Ti, GPU-test, 591.74, 16311, 5260, 100, 72, 180.29, 84, 2602, 16001, P0, 3, 8\n")

	sample, err := parseGPUSample(output, basicGPUQueryFields)
	if err != nil {
		t.Fatal(err)
	}
	if len(sample.GPUs) != 1 {
		t.Fatalf("expected one GPU, got %d", len(sample.GPUs))
	}
	gpu := sample.GPUs[0]
	if gpu.MemoryTotalBytes != 16311*1024*1024 || gpu.MemoryUsedBytes != 5260*1024*1024 {
		t.Fatalf("unexpected memory totals %d/%d", gpu.MemoryUsedBytes, gpu.MemoryTotalBytes)
	}
	if gpu.PCIeLinkGeneration != "3" || gpu.PCIeLinkWidth != "8" {
		t.Fatalf("unexpected PCIe link %q x%s", gpu.PCIeLinkGeneration, gpu.PCIeLinkWidth)
	}
	if gpu.VBIOSVersion != "" || gpu.PowerLimitWatts != nil {
		t.Fatalf("expected rich-only fields to be empty")
	}
}
