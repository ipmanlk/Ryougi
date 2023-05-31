package main

import (
	"fmt"
	"ipmanlk/saika/anilist"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	defer logMemoryUsage()

	searchText := "Naruto"
	media, err := anilist.SearchMedia(searchText, "ANIME")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("%s", media[0].Hash())

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}

func logMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Conversion factors
	const (
		_  = 1 << (10 * iota)
		KB // 1024
		MB // 1024 * 1024
		GB // 1024 * 1024 * 1024
	)

	// Calculate memory usage in MB and GB
	allocMB := float64(memStats.Alloc) / MB
	totalAllocGB := float64(memStats.TotalAlloc) / GB
	sysGB := float64(memStats.Sys) / GB

	mallocs := memStats.Mallocs
	frees := memStats.Frees

	fmt.Printf("Alloc: %.2f MB; TotalAlloc: %.2f GB; Sys: %.2f GB; Mallocs: %d; Frees: %d\n",
		allocMB, totalAllocGB, sysGB, mallocs, frees)
}
