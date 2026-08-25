package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"fuelhydrant/internal/alarm"
	"fuelhydrant/internal/hydrant"
	"fuelhydrant/internal/model"
	"fuelhydrant/internal/pump"
	"fuelhydrant/internal/record"
	"fuelhydrant/internal/supply"
)

type pumpScheduleEntry struct {
	PumpID   string `json:"id"`
	Priority int    `json:"priority"`
	OnDemand bool   `json:"onDemand"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *service) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/console.html")
}

func (s *service) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleSupply(w http.ResponseWriter, r *http.Request) {
	state := s.supply.State()
	writeJSON(w, http.StatusOK, map[string]any{
		"state":     state,
		"fuel":      s.supply.Fuel(),
		"idle":      state.IsIdle(),
		"supplying": state.IsSupplying(),
	})
}

func (s *service) handleSupplyStart(w http.ResponseWriter, r *http.Request) {
	if err := s.supply.Start(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"state": string(s.supply.State())})
}

func (s *service) handleSupplyStop(w http.ResponseWriter, r *http.Request) {
	if err := s.supply.Stop(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"state": string(s.supply.State())})
}

func (s *service) handleSupplySwitch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Fuel string `json:"fuel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.supply.SwitchFuel(model.FuelType(body.Fuel)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"fuel": string(s.supply.Fuel())})
}

func (s *service) handleSupplyTarget(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Target    float64 `json:"target"`
		Tolerance float64 `json:"tolerance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.supply.SetTarget(body.Target, body.Tolerance)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleSupplySequence(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Fuel    string   `json:"fuel"`
		PumpIDs []string `json:"pumpIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.supply.ApplySequence(supply.Sequence{Fuel: model.FuelType(body.Fuel), PumpIDs: body.PumpIDs})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleSupplyDispatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Fuel string `json:"fuel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.supply.DispatchSequence(model.FuelType(body.Fuel))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handlePumps(w http.ResponseWriter, r *http.Request) {
	type item struct {
		ID      string `json:"id"`
		State   string `json:"state"`
		Mode    string `json:"mode"`
		Failed  bool   `json:"failed"`
		Retries int    `json:"retries"`
		Running bool   `json:"running"`
	}
	items := make([]item, 0)
	for _, id := range s.group.IDs() {
		it := item{ID: id, State: string(s.group.State(id)), Mode: string(s.group.Mode())}
		if p, ok := s.group.Pump(id); ok {
			it.Failed = p.Failed()
			it.Retries = p.Retries()
			it.Running = p.State() == model.PumpRunning
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"mode":       s.group.Mode(),
		"manual":     s.group.IsManual(),
		"anyRunning": s.group.IsRunning(),
		"pumps":      items,
	})
}

func (s *service) handlePumpMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.group.SetMode(model.PumpMode(body.Mode))
	writeJSON(w, http.StatusOK, map[string]string{"mode": string(s.group.Mode())})
}

func (s *service) handlePumpManual(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID    string `json:"id"`
		Start bool   `json:"start"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var err error
	if body.Start {
		err = s.group.ManualStart(body.ID)
	} else {
		err = s.group.ManualStop(body.ID)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handlePumpHealth(w http.ResponseWriter, r *http.Request) {
	motors := make(map[string]bool)
	hydrantPressure := make(map[string]float64)
	for _, id := range s.group.IDs() {
		motors[id] = s.group.MotorRunning(id)
	}
	for _, st := range s.monitor.List() {
		if pressure, ok := s.group.HydrantPressure(st.ID); ok {
			hydrantPressure[st.ID] = pressure
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"health":          s.group.Health(),
		"running":         s.group.RunningCount(),
		"summary":         s.group.Summary(),
		"anyRunning":      s.group.IsRunning(),
		"motorRunning":    motors,
		"hydrantPressure": hydrantPressure,
	})
}

func (s *service) handleHydrants(w http.ResponseWriter, r *http.Request) {
	type item struct {
		ID            string  `json:"id"`
		Pressure      float64 `json:"pressure"`
		Flow          float64 `json:"flow"`
		Leaking       bool    `json:"leaking"`
		Fuel          string  `json:"fuel"`
		Known         bool    `json:"known"`
		PressureValue float64 `json:"pressureValue"`
		FlowValue     float64 `json:"flowValue"`
		FlowRate      float64 `json:"flowRate"`
	}
	items := make([]item, 0, len(s.monitor.List()))
	for _, st := range s.monitor.List() {
		it := item{
			ID:       st.ID,
			Pressure: st.Pressure,
			Flow:     st.Flow,
			Leaking:  st.Leaking,
			Fuel:     string(st.Fuel),
			Known:    s.monitor.Known(st.ID),
		}
		if pressure, err := s.supply.HydrantPressure(st.ID); err == nil {
			it.Pressure = pressure
		}
		if h := s.monitor.HydrantPtr(st.ID); h != nil {
			it.PressureValue = h.PressureValue()
			it.FlowValue = h.FlowValue()
		}
		if rate, ok := s.monitor.FlowRate(st.ID); ok {
			it.FlowRate = rate
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *service) handleHydrantMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"average_pressure": s.monitor.AveragePressure(),
		"total_flow":       s.monitor.TotalFlow(),
		"leaking_count":    s.monitor.LeakingCount(),
		"hydrant_count":    s.monitor.Count(),
	})
}

func (s *service) handleHydrantRecover(w http.ResponseWriter, r *http.Request) {
	s.monitor.CacheSnapshot()
	s.monitor.Recover(s.monitor.Snapshot())
	for _, st := range s.monitor.List() {
		s.monitor.RefreshCache(st.ID)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleBelowPressure(w http.ResponseWriter, r *http.Request) {
	threshold := 2.5
	if raw := r.URL.Query().Get("threshold"); raw != "" {
		_, _ = fmt.Sscanf(raw, "%f", &threshold)
	}
	thresholds := hydrant.DefaultThresholds()
	below := s.monitor.BelowPressure(threshold)
	lowPressure := make([]string, 0)
	lowFlow := make([]string, 0)
	withinRange := make([]string, 0)
	for _, st := range s.monitor.List() {
		lp, lf := s.monitor.Evaluate(st.ID, thresholds)
		if lp {
			lowPressure = append(lowPressure, st.ID)
		}
		if lf {
			lowFlow = append(lowFlow, st.ID)
		}
		if pressure, ok := s.monitor.PressureOf(st.ID); ok && model.WithinRange(pressure, thresholds.LowPressure, thresholds.HighPressure) {
			withinRange = append(withinRange, st.ID)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"below":       below,
		"lowPressure": lowPressure,
		"lowFlow":     lowFlow,
		"withinRange": withinRange,
		"defaultLow":  thresholds.LowPressure,
		"defaultHigh": thresholds.HighPressure,
	})
}

func (s *service) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	var body model.Telemetry
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.monitor.ApplyTelemetry(body)
	s.group.SyncHydrantPressure(body.Hydrant, body.Pressure)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleTelemetryBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []model.Telemetry `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.supply.ApplyTelemetryBatch(body.Items)
	s.monitor.CacheSnapshot()
	writeJSON(w, http.StatusOK, map[string]int{"applied": len(body.Items)})
}

func (s *service) handleAlarms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.alarm.List())
}

func (s *service) handleActiveAlarms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"active": s.alarm.ActiveAlarms(),
		"total":  s.alarm.TotalAlarms(),
		"count":  s.alarm.Count(),
	})
}

func (s *service) handleAlarmHistory(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.alarm.History())
}

func (s *service) handleAlarmRaise(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID   string `json:"id"`
		Kind string `json:"kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if body.ID == "" || body.Kind == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id and kind are required"})
		return
	}
	s.alarm.Raise(body.ID, body.Kind)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"acknowledged": s.alarm.Acknowledged(body.ID),
	})
}

func (s *service) handleAlarmEscalate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	level := s.alarm.Escalate(body.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"level":      level,
		"escalation": alarm.Escalation{ID: body.ID, Level: level},
	})
}

func (s *service) handleAlarmAck(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !s.alarm.Acknowledge(body.ID) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "alarm not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"acknowledged": s.alarm.Acknowledged(body.ID),
	})
}

func (s *service) handleLeakDetect(w http.ResponseWriter, r *http.Request) {
	found := s.leak.Detect(s.monitor.List())
	s.alarm.Reconcile(s.monitor.List())
	writeJSON(w, http.StatusOK, map[string]any{"leak": found, "interlock": s.leak.InterlockState()})
}

func (s *service) handleLeakRecover(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.monitor.SetLeaking(body.ID, false)
	if err := s.alarm.Clear(body.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.alarm.ResolveHydrant(body.ID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleLeakStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"interlock": s.leak.InterlockState(),
		"active":    s.leak.ActiveLeaks(),
		"assessed":  s.leak.EvaluateStatus(s.monitor.List(), 2.5),
	})
}

func (s *service) handleInterlockReset(w http.ResponseWriter, r *http.Request) {
	s.leak.ResetInterlock()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handlePressureDrop(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Previous  map[string]float64 `json:"previous"`
		Current   map[string]float64 `json:"current"`
		Threshold float64            `json:"threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, s.leak.DetectPressureDrops(body.Previous, body.Current, body.Threshold))
}

func (s *service) handleRecord(w http.ResponseWriter, r *http.Request) {
	offset := 0
	limit := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &offset)
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
	}
	events, err := s.journal.Page(offset, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *service) handleRecordStats(w http.ResponseWriter, r *http.Request) {
	counts, err := s.journal.CountByKind()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	digest, err := s.journal.Integrity()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"counts":      counts,
		"events":      digest.Events,
		"sum":         digest.Sum,
		"path":        s.journal.Path(),
		"openHandles": s.journal.OpenHandles(),
		"closed":      s.journal.Closed(),
		"kinds":       record.SortedKinds(counts),
	})
}

func (s *service) handleRecordExport(w http.ResponseWriter, r *http.Request) {
	payload, err := s.journal.Export()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(payload))
}

func (s *service) handleRecordRotate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.journal.Rotate(body.Path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleSupplyMetrics(w http.ResponseWriter, r *http.Request) {
	target, tolerance := s.supply.PressureTarget()
	writeJSON(w, http.StatusOK, map[string]any{
		"average_pressure": s.supply.AveragePressure(),
		"consumption":      s.supply.ConsumptionRate(),
		"target":           target,
		"tolerance":        tolerance,
		"flush_cycles":     s.supply.FlushCycles(),
		"last_fuel":        s.supply.LastFuel(),
	})
}

func (s *service) handleSwitchStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"state":       s.switcher.State(),
		"from":        s.switcher.From(),
		"to":          s.switcher.To(),
		"drain":       s.supply.SwitchStatus()["drain"],
		"in_progress": s.switcher.InProgress(),
	})
}

func (s *service) handleSwitchReset(w http.ResponseWriter, r *http.Request) {
	s.switcher.Reset()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handleDrainProgress(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Progress float64 `json:"progress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.supply.SetDrainProgress(body.Progress)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handlePumpFailover(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Primary string `json:"primary"`
		Backup  string `json:"backup"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.group.Failover(body.Primary, body.Backup); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *service) handlePumpSchedule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Entries []pumpScheduleEntry `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	entries := make([]pump.ScheduleEntry, 0, len(body.Entries))
	for _, entry := range body.Entries {
		entries = append(entries, pump.ScheduleEntry{
			PumpID:   entry.PumpID,
			Priority: entry.Priority,
			OnDemand: entry.OnDemand,
		})
	}
	writeJSON(w, http.StatusOK, s.group.Schedule(entries))
}
