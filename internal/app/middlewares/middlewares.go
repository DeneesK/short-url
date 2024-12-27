package middlewares

import (
	"log"
	"net/http"

	"github.com/DeneesK/short-url/internal/pkg/memmetrics"
)

const (
	RAM  = "ram"
	Disk = "disk"
	Both = "both"
)

// MemoryControl monitors the amount of used memory and interrupts
// the execution of new requests if the limit is exceeded.
//
// memoryUsageLimit maximum memory available for recording,
// the value is set in percentage. memoryUsageLimit must be 0 => n <= 100.
//
// memoryType: "ram", "disk", "both"
func NewMemoryControlMiddleware(memoryUsageLimit float64, memoryType string) func(next http.Handler) http.Handler {

	if memoryUsageLimit > 100 || memoryUsageLimit < 0 {
		log.Fatal("NewMemoryControlMiddleware: memoryUsageLimit must be 0 => n <= 100")
	} else if !(isValidMemType(memoryType)) {
		log.Fatal("NewMemoryControlMiddleware: memoryType must be ram, disk or both")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch memoryType {
			case RAM:
				ramStat, err := memmetrics.CurrentRAMStat()
				if err != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				if memoryUsageLimit < (float64(ramStat.FreeRAM)/float64(ramStat.TotalRAM))*100 {
					http.Error(w, "memory limit is exceeded", http.StatusInternalServerError)
					return
				}
			case Disk:
				diskSpaceStat, err := memmetrics.CurrentDiskSpaceStat()
				if err != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				if memoryUsageLimit < (float64(diskSpaceStat.FreeDiskSpace)/float64(diskSpaceStat.TotalDiskSpace))*100 {
					http.Error(w, "memory limit is exceeded", http.StatusInternalServerError)
					return
				}
			case Both:
				memStat, err := memmetrics.CurrentMemoryStat()
				if err != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				if memoryUsageLimit < (float64(memStat.RAMStat.FreeRAM)/float64(memStat.RAMStat.TotalRAM))*100 {
					http.Error(w, "memory limit is exceeded", http.StatusInternalServerError)
					return
				} else if memoryUsageLimit < (float64(memStat.DiskSpaceStat.FreeDiskSpace)/float64(memStat.DiskSpaceStat.TotalDiskSpace))*100 {
					http.Error(w, "memory limit is exceeded", http.StatusInternalServerError)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isValidMemType(memType string) bool {
	switch memType {
	case "ram", "disk", "both":
		return true
	default:
		return false
	}
}
