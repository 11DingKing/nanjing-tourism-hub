package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func TestWarningIncidentAndEscalationWorkflow(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-escalation")
	if incident.Status != domain.IncidentActive {
		t.Fatalf("unexpected incident status %s", incident.Status)
	}
	response, err := env.service.EscalateResponse(context.Background(), domain.EscalateResponseCommand{
		IncidentID:  incident.ID,
		RegionID:    "coast-a",
		FromVersion: 1,
		Target:      domain.ResponseIII,
		Reason:      "累计雨量接近极端阈值",
		Actor:       env.dispatcher,
	})
	if err != nil {
		t.Fatalf("escalate response: %v", err)
	}
	if response.Level != domain.ResponseIII {
		t.Fatalf("got response level %s", response.Level)
	}
	if response.Version != 2 {
		t.Fatalf("got response version %d", response.Version)
	}
	_, err = env.service.EscalateResponse(context.Background(), domain.EscalateResponseCommand{
		IncidentID:  incident.ID,
		RegionID:    "coast-a",
		FromVersion: 1,
		Target:      domain.ResponseII,
		Reason:      "stale command",
		Actor:       env.dispatcher,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
}

func TestShelterReservationAndReleasePreserveCapacity(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-shelter")
	reservation, err := env.service.ReserveShelter(context.Background(), domain.ReserveShelterCommand{
		ReservationID:  "reservation-1",
		ShelterID:      "shelter-1",
		IncidentID:     incident.ID,
		RegionID:       "coast-a",
		IdempotencyKey: "evacuate-village-a",
		People:         75,
		ExpiresAt:      env.now.Add(4 * time.Hour),
		Actor:          env.dispatcher,
	})
	if err != nil {
		t.Fatalf("reserve shelter: %v", err)
	}
	_, err = env.service.ReserveShelter(context.Background(), domain.ReserveShelterCommand{
		ReservationID:  "reservation-2",
		ShelterID:      "shelter-1",
		IncidentID:     incident.ID,
		RegionID:       "coast-a",
		IdempotencyKey: "evacuate-village-b",
		People:         30,
		ExpiresAt:      env.now.Add(4 * time.Hour),
		Actor:          env.dispatcher,
	})
	if !errors.Is(err, domain.ErrCapacity) {
		t.Fatalf("expected capacity error, got %v", err)
	}
	err = env.service.ReleaseShelter(context.Background(), domain.ReleaseShelterCommand{
		ReservationID:   reservation.ID,
		ExpectedVersion: reservation.Version,
		Actor:           env.dispatcher,
	})
	if err != nil {
		t.Fatalf("release shelter: %v", err)
	}
	_, err = env.service.ReserveShelter(context.Background(), domain.ReserveShelterCommand{
		ReservationID:  "reservation-3",
		ShelterID:      "shelter-1",
		IncidentID:     incident.ID,
		RegionID:       "coast-a",
		IdempotencyKey: "evacuate-village-c",
		People:         100,
		ExpiresAt:      env.now.Add(4 * time.Hour),
		Actor:          env.dispatcher,
	})
	if err != nil {
		t.Fatalf("capacity was not released: %v", err)
	}
}

func TestConcurrentShelterReservationsDoNotOversubscribe(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-concurrent-shelter")
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for index := 0; index < 2; index++ {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := env.service.ReserveShelter(context.Background(), domain.ReserveShelterCommand{
				ReservationID:  "concurrent-reservation-" + string(rune('a'+index)),
				ShelterID:      "shelter-1",
				IncidentID:     incident.ID,
				RegionID:       "coast-a",
				IdempotencyKey: "concurrent-key-" + string(rune('a'+index)),
				People:         60,
				ExpiresAt:      env.now.Add(time.Hour),
				Actor:          env.dispatcher,
			})
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	succeeded := 0
	failed := 0
	for err := range results {
		if err == nil {
			succeeded++
		} else {
			failed++
		}
	}
	if succeeded != 1 || failed != 1 {
		t.Fatalf("expected one reservation to win, succeeded=%d failed=%d", succeeded, failed)
	}
}

func TestDispatchHandoffAndTaskOwnershipMoveAtomically(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-dispatch")
	task := env.seedTask(t, incident.ID, "task-dispatch")
	dispatch, err := env.service.DispatchTeam(context.Background(), domain.DispatchTeamCommand{
		DispatchID:  "dispatch-1",
		TeamID:      "team-1",
		IncidentID:  incident.ID,
		RegionID:    "coast-a",
		TaskID:      task.ID,
		TeamVersion: 1,
		Actor:       env.dispatcher,
	})
	if err != nil {
		t.Fatalf("dispatch team: %v", err)
	}
	if dispatch.TeamID != "team-1" {
		t.Fatalf("unexpected dispatch team %s", dispatch.TeamID)
	}
	changed, err := env.service.HandoffTeam(context.Background(), domain.HandoffTeamCommand{
		DispatchID:  "dispatch-1",
		FromTeamID:  "team-1",
		ToTeamID:    "team-2",
		FromVersion: 2,
		ToVersion:   1,
		Actor:       env.dispatcher,
	})
	if err != nil {
		t.Fatalf("handoff team: %v", err)
	}
	if changed.TeamID != "team-2" {
		t.Fatalf("handoff retained old team %s", changed.TeamID)
	}
}

func TestSupplyAllocationCancelAndReuse(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-supply")
	task := env.seedTask(t, incident.ID, "task-supply")
	allocation, err := env.service.AllocateSupply(context.Background(), domain.AllocateSupplyCommand{
		AllocationID:   "allocation-1",
		LotID:          "lot-water",
		IncidentID:     incident.ID,
		RegionID:       "coast-a",
		TaskID:         task.ID,
		IdempotencyKey: "water-village-a",
		Quantity:       160,
		LotVersion:     1,
		Actor:          env.dispatcher,
	})
	if err != nil {
		t.Fatalf("allocate supply: %v", err)
	}
	_, err = env.service.AllocateSupply(context.Background(), domain.AllocateSupplyCommand{
		AllocationID:   "allocation-2",
		LotID:          "lot-water",
		IncidentID:     incident.ID,
		RegionID:       "coast-a",
		TaskID:         task.ID,
		IdempotencyKey: "water-village-b",
		Quantity:       50,
		LotVersion:     2,
		Actor:          env.dispatcher,
	})
	if !errors.Is(err, domain.ErrCapacity) {
		t.Fatalf("expected capacity error, got %v", err)
	}
	if err := env.service.CancelAllocation(context.Background(), domain.CancelAllocationCommand{
		AllocationID:      allocation.ID,
		AllocationVersion: allocation.Version,
		Actor:             env.dispatcher,
	}); err != nil {
		t.Fatalf("cancel allocation: %v", err)
	}
	_, err = env.service.AllocateSupply(context.Background(), domain.AllocateSupplyCommand{
		AllocationID:   "allocation-3",
		LotID:          "lot-water",
		IncidentID:     incident.ID,
		RegionID:       "coast-a",
		TaskID:         task.ID,
		IdempotencyKey: "water-village-c",
		Quantity:       200,
		LotVersion:     3,
		Actor:          env.dispatcher,
	})
	if err != nil {
		t.Fatalf("canceled quantity was not restored: %v", err)
	}
}

func TestReceiptRejectionReturnsTaskToExecution(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-receipt")
	task := env.seedTask(t, incident.ID, "task-receipt")
	accepted, err := env.service.AdvanceTask(context.Background(), domain.AdvanceTaskCommand{
		TaskID:          task.ID,
		ExpectedVersion: task.Version,
		Target:          domain.TaskAccepted,
		Actor:           env.dispatcher,
	})
	if err != nil {
		t.Fatalf("accept task: %v", err)
	}
	executing, err := env.service.AdvanceTask(context.Background(), domain.AdvanceTaskCommand{
		TaskID:          task.ID,
		ExpectedVersion: accepted.Version,
		Target:          domain.TaskExecuting,
		Actor:           env.dispatcher,
	})
	if err != nil {
		t.Fatalf("start task: %v", err)
	}
	receipt, err := env.service.SubmitReceipt(context.Background(), domain.SubmitReceiptCommand{
		Receipt: domain.Receipt{
			ID:            "receipt-1",
			TaskID:        task.ID,
			IncidentID:    incident.ID,
			RegionID:      "coast-a",
			Summary:       "移动泵部署完毕，低洼点水位回落",
			EvidenceCount: 2,
		},
		TaskVersion: executing.Version,
		Actor:       env.field,
	})
	if err != nil {
		t.Fatalf("submit receipt: %v", err)
	}
	reviewed, err := env.service.ReviewReceipt(context.Background(), domain.ReviewReceiptCommand{
		ReceiptID:       receipt.ID,
		ExpectedVersion: receipt.Version,
		Accept:          false,
		Note:            "缺少积水深度复测数据",
		Actor:           env.reviewer,
	})
	if err != nil {
		t.Fatalf("review receipt: %v", err)
	}
	if reviewed.Status != "rejected" {
		t.Fatalf("unexpected receipt status %s", reviewed.Status)
	}
	page, err := env.service.ListTasks(context.Background(), domain.TaskFilter{
		IncidentID: incident.ID,
		Page:       domain.Page{Limit: 20, Sort: "created_at", Direction: "asc"},
	}, env.duty)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Status != domain.TaskExecuting {
		t.Fatalf("task was not reopened: %#v", page.Items)
	}
}

func TestCanceledBatchReturnsCompletedPrefix(t *testing.T) {
	env := newTestEnvironment(t)
	incident := env.seedIncident(t, "incident-batch")
	task := env.seedTask(t, incident.ID, "task-batch")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := env.service.BatchAssign(ctx, domain.BatchAssignCommand{
		IncidentID: incident.ID,
		Items: []domain.BatchAssignItem{
			{TaskID: task.ID, TeamID: "team-1", TaskVersion: task.Version, TeamVersion: 1},
		},
		Actor: env.dispatcher,
	})
	if !errors.Is(err, domain.ErrCanceled) {
		t.Fatalf("expected cancellation, got result=%#v err=%v", result, err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("canceled batch reported unprocessed entries: %#v", result.Items)
	}
}
