package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ryan/wvwlog/analysis"
	"github.com/ryan/wvwlog/config"
	"github.com/ryan/wvwlog/parser"
	"github.com/ryan/wvwlog/processor"
)

//go:embed all:dist
var distFS embed.FS

// ── Fight cache ───────────────────────────────────────────────────────────────

// FightMeta is a lightweight summary used in the sidebar fight list.
type FightMeta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DurationMs  int    `json:"durationMs"`
	PlayerCount int    `json:"playerCount"`
}

// FightDetail is the full analysis payload returned for a single fight.
type FightDetail struct {
	Meta          FightMeta                          `json:"meta"`
	Players       []analysis.PlayerSummary           `json:"players"`
	WellTiming    map[int]*analysis.WellSyncAnalysis `json:"wellTiming"`
	SpikeDamage   map[string][]analysis.SpikeDamage  `json:"spikeDamage"`
	SyncedOverlap []analysis.SyncedSpikeWell         `json:"syncedOverlap"`
}

var (
	fightsMu sync.RWMutex
	fights   = []FightMeta{}
	details  = map[string]*FightDetail{}
)

// ── Log state ─────────────────────────────────────────────────────────────────

// LogStatus represents the processing state of a .zevtc log file.
type LogStatus string

const (
	LogPending    LogStatus = "pending"
	LogProcessing LogStatus = "processing"
	LogDone       LogStatus = "done"
	LogError      LogStatus = "error"
)

// LogEntry is the per-file state exposed via /api/logs.
type LogEntry struct {
	Name   string    `json:"name"`
	Status LogStatus `json:"status"`
	Error  string    `json:"error,omitempty"`
}

var (
	logsMu   sync.RWMutex
	logState = map[string]*LogEntry{} // key = filename e.g. "20260308-090913.zevtc"
	procSem  = make(chan struct{}, max(runtime.NumCPU()/2, 2)) // scale with CPUs, min 2
)

// ── Startup ───────────────────────────────────────────────────────────────────

// LoadFights registers all .zevtc files in LogFolder as pending.
func LoadFights() {
	scanLogFolder(false)
}

// scanLogFolder scans LogFolder for .zevtc files. Files not yet tracked are
// registered as pending. If autoProcess=true, newly discovered files are also
// immediately queued for processing (used by the watcher for new arrivals).
func scanLogFolder(autoProcess bool) {
	entries, err := os.ReadDir(config.AppConfig.LogFolder)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".zevtc" {
			continue
		}
		name := e.Name()
		logsMu.Lock()
		_, known := logState[name]
		if !known {
			logState[name] = &LogEntry{Name: name, Status: LogPending}
		}
		logsMu.Unlock()
		if !known && autoProcess {
			go processLog(name)
		}
	}
}

// StartWatcher polls LogFolder every 3 seconds and auto-processes any new
// .zevtc files that weren't present at startup.
func StartWatcher() {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			scanLogFolder(true)
		}
	}()
}

// ── Processing ────────────────────────────────────────────────────────────────

func processLog(name string) {
	logsMu.Lock()
	e, ok := logState[name]
	if !ok || e.Status == LogProcessing || e.Status == LogDone {
		logsMu.Unlock()
		return
	}
	e.Status = LogProcessing
	e.Error = ""
	logsMu.Unlock()

	procSem <- struct{}{}        // acquire
	defer func() { <-procSem }() // release

	fullPath := filepath.Join(config.AppConfig.LogFolder, name)
	jsonPath, err := processor.ConvertLog(config.AppConfig.EliteInsightsCLI, fullPath)
	if err != nil {
		logsMu.Lock()
		logState[name].Status = LogError
		logState[name].Error = err.Error()
		logsMu.Unlock()
		log.Printf("convert error %s: %v", name, err)
		return
	}

	if err := importFightFromJSON(jsonPath); err != nil {
		logsMu.Lock()
		logState[name].Status = LogError
		logState[name].Error = err.Error()
		logsMu.Unlock()
		log.Printf("parse error %s: %v", name, err)
		os.Remove(jsonPath)
		return
	}

	logsMu.Lock()
	logState[name].Status = LogDone
	logsMu.Unlock()
}

func importFightFromJSON(path string) error {
	fight, err := parser.ParseFight(path)
	if err != nil {
		return err
	}

	id := strings.TrimSuffix(filepath.Base(path), ".json")
	meta := FightMeta{
		ID:          id,
		Name:        fight.Name,
		DurationMs:  fight.Duration,
		PlayerCount: len(fight.Players),
	}

	players := analysis.BuildFightSummary(fight)

	// Run analyses concurrently.
	var wg sync.WaitGroup
	var wellTiming map[int]*analysis.WellSyncAnalysis
	var spikeDamage map[string][]analysis.SpikeDamage

	wg.Add(2)
	go func() {
		defer wg.Done()
		wellTiming = analysis.AnalyzeWellSkillTiming(fight)
	}()
	go func() {
		defer wg.Done()
		spikeDamage = analysis.AnalyzeSpikedDamage(
			fight, 5, 0,
			config.AppConfig.SpikeIncreaseThreshold,
			config.AppConfig.MinSpikeDPS,
		)
	}()
	wg.Wait()

	syncedOverlap := analysis.FindSyncedSpikeWells(wellTiming, spikeDamage)

	d := &FightDetail{
		Meta:          meta,
		Players:       players,
		WellTiming:    wellTiming,
		SpikeDamage:   spikeDamage,
		SyncedOverlap: syncedOverlap,
	}

	fightsMu.Lock()
	details[id] = d
	found := false
	for i, f := range fights {
		if f.ID == id {
			fights[i] = meta
			found = true
			break
		}
	}
	if !found {
		fights = append(fights, meta)
		sort.Slice(fights, func(i, j int) bool { return fights[i].ID < fights[j].ID })
	}
	fightsMu.Unlock()

	if err := os.Remove(path); err != nil {
		log.Printf("warning: failed to delete %s: %v", filepath.Base(path), err)
	}
	log.Printf("imported fight: %s (%s, %d players)", id, fight.Name, len(fight.Players))
	return nil
}

// ── HTTP server ───────────────────────────────────────────────────────────────

// Start starts the HTTP dashboard server on the given address (e.g. ":8080").
func Start(addr string) error {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		return fmt.Errorf("embed sub: %w", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/logs", handleLogs)
	mux.HandleFunc("/api/logs/process", handleProcessLogs)
	mux.HandleFunc("/api/fights", handleFightList)
	mux.HandleFunc("/api/fights/", handleFightDetail)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := subFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	fmt.Printf("\nDashboard: http://localhost%s\n\n", addr)
	return http.ListenAndServe(addr, mux)
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func handleLogs(w http.ResponseWriter, r *http.Request) {
	logsMu.RLock()
	list := make([]*LogEntry, 0, len(logState))
	for _, e := range logState {
		copy := *e
		list = append(list, &copy)
	}
	logsMu.RUnlock()

	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type processRequest struct {
	Names []string `json:"names"`
}

func handleProcessLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	for _, name := range req.Names {
		go processLog(name)
	}
	w.WriteHeader(http.StatusAccepted)
}

func handleFightList(w http.ResponseWriter, r *http.Request) {
	fightsMu.RLock()
	list := fights
	fightsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func handleFightDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/fights/")
	if id == "" {
		http.Error(w, "missing fight id", http.StatusBadRequest)
		return
	}
	fightsMu.RLock()
	d, ok := details[id]
	fightsMu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}
