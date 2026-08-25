package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fuelhydrant/internal/alarm"
	"fuelhydrant/internal/hydrant"
	"fuelhydrant/internal/leak"
	"fuelhydrant/internal/model"
	"fuelhydrant/internal/pump"
	"fuelhydrant/internal/record"
	"fuelhydrant/internal/supply"
	"fuelhydrant/internal/switch"
)

type service struct {
	monitor  *hydrant.Monitor
	group    *pump.Group
	alarm    *alarm.Manager
	leak     *leak.Detector
	supply   *supply.Coordinator
	switcher *fuelswitch.State
	journal  *record.Journal
}

func newService(dir string) *service {
	_ = os.MkdirAll(dir, 0755)
	journal := record.NewJournal(filepath.Join(dir, "run.log"))
	_ = journal.Open()

	leakDetector := leak.NewDetector()
	alarmManager := alarm.NewManager(journal, leakDetector)
	leakDetector.SetRaiseAlarm(func(id string) { alarmManager.RaiseHydrantLeak(id) })
	leakDetector.SetWriteback(func(id string) error { return journal.Append("leak-restore", id) })

	group := pump.NewGroup()
	leakDetector.SetStopPumps(func() error { return group.StopAll() })

	for _, id := range []string{"pump-1", "pump-2"} {
		p := pump.NewPump(id, pump.NewMotor(), alarmManager)
		group.AddPump(p)
	}

	monitor := hydrant.NewMonitor()
	for _, id := range []string{"hydrant-1", "hydrant-2", "hydrant-3"} {
		monitor.Register(id, model.FuelJetA1)
	}

	switcher := fuelswitch.New()
	supplyCoordinator := supply.NewCoordinator(monitor, group, switcher)

	return &service{
		monitor:  monitor,
		group:    group,
		alarm:    alarmManager,
		leak:     leakDetector,
		supply:   supplyCoordinator,
		switcher: switcher,
		journal:  journal,
	}
}

func (s *service) runSelfCheck() {
	if err := s.journal.Append("selfcheck", "start"); err != nil {
		fmt.Println("selfcheck journal failed:", err)
		os.Exit(1)
	}
	s.monitor.ApplyTelemetry(model.Telemetry{Hydrant: "hydrant-1", Pressure: 4.1, Flow: 2.0})
	pressure, ok := s.monitor.Pressure("hydrant-1")
	if !ok || pressure != 4.1 {
		fmt.Println("selfcheck telemetry failed")
		os.Exit(1)
	}
	s.group.SetPressure(4.1)
	if !s.group.PressureStable(4.0, 0.2) {
		fmt.Println("selfcheck pressure failed")
		os.Exit(1)
	}
	if err := s.group.StartAll(); err != nil {
		fmt.Println("selfcheck pump start failed:", err)
		os.Exit(1)
	}
	if s.group.RunningCount() != 2 {
		fmt.Println("selfcheck running count failed")
		os.Exit(1)
	}
	_ = s.group.StopAll()
	s.monitor.SetLeaking("hydrant-1", true)
	s.leak.Detect(s.monitor.List())
	if s.leak.InterlockState() != model.InterlockLocked {
		fmt.Println("selfcheck interlock failed")
		os.Exit(1)
	}
	s.monitor.SetLeaking("hydrant-1", false)
	if err := s.alarm.Clear("hydrant-1"); err != nil {
		fmt.Println("selfcheck restore failed:", err)
		os.Exit(1)
	}
	s.supply.SetDrainProgress(1.0)
	digest, err := s.journal.Integrity()
	if err != nil || digest.Events == 0 {
		fmt.Println("selfcheck integrity failed")
		os.Exit(1)
	}
	_, err = s.journal.ChecksumOf("alarm")
	if err != nil {
		fmt.Println("selfcheck checksum failed:", err)
		os.Exit(1)
	}
	fmt.Println("selfcheck ok")
}
