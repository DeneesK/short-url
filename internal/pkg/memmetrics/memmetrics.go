package memmetrics

import (
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type RAMStat struct {
	TotalRAM uint64
	FreeRAM  uint64
}

type DiskSpaceStat struct {
	TotalDiskSpace uint64
	FreeDiskSpace  uint64
}

type MemoryStat struct {
	RAMStat       RAMStat
	DiskSpaceStat DiskSpaceStat
}

func CurrentRAMStat() (RAMStat, error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return RAMStat{}, err
	}
	return RAMStat{TotalRAM: memStat.Total, FreeRAM: memStat.Available}, nil
}

func CurrentDiskSpaceStat() (DiskSpaceStat, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return DiskSpaceStat{}, err
	}
	return DiskSpaceStat{TotalDiskSpace: diskStat.Total, FreeDiskSpace: diskStat.Free}, nil
}

func CurrentMemoryStat() (MemoryStat, error) {
	d, err := CurrentDiskSpaceStat()
	if err != nil {
		return MemoryStat{}, err
	}

	r, err := CurrentRAMStat()
	if err != nil {
		return MemoryStat{}, err
	}
	return MemoryStat{RAMStat: r, DiskSpaceStat: d}, nil
}
