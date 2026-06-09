package recovery

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	agentpb "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"
)

// StreamProvider is the interface the dispatcher uses to send commands to agents
// and check online status. The concrete implementation lives in the agent package
// (Phase 3). Defined here to avoid import cycles: recovery ← agent would cycle
// because agent_imp.go will import recovery for hooks in Phase 3.
type StreamProvider interface {
	// GetStream returns the bidirectional stream for the agent, if connected.
	GetStream(agentID uint) (AgentStream, bool)
	// IsOnline returns true if the agent is considered online (last ping < 1m).
	IsOnline(agentID uint) bool
}

// AgentStream is the minimum interface needed to send a command to an agent.
// The concrete type is agent.AgentService_AgentStreamServer from the gRPC package.
type AgentStream interface {
	Send(*agentpb.BidirectionalStream) error
}

// dispatcherState holds all mutable state for the recovery dispatch engine.
// A single package-level instance is used (dispatcher).
type dispatcherState struct {
	started      atomic.Bool
	shuttingDown atomic.Bool

	mu     sync.Mutex // guards ctx/cancel/ticker/db/provider
	ctx    context.Context
	cancel context.CancelFunc
	db     *gorm.DB
	ticker *time.Ticker

	wg sync.WaitGroup // tracks in-flight dispatch goroutines

	// pendingResults maps cmd_id → target ID for ACK matching.
	pendingResults sync.Map // map[string]uint

	// globalSem caps total in-flight dispatches across all recoveries.
	globalSem *semaphore.Weighted

	// perRecoverySem caps in-flight per recovery_id. Lazily created.
	perRecoverySem sync.Map // map[uint]*semaphore.Weighted

	// provider is the StreamProvider set via SetStreamProvider / Start.
	provider StreamProvider

	// dependencyWaitLogged deduplicates recovery_dependency_wait log emissions
	// within a single evaluation cycle. The pointer is REPLACED at the start
	// of each safety-net tick so concurrent dispatches racing with the reset
	// cannot lose or duplicate entries (Range+Delete is non-atomic and races
	// with concurrent LoadOrStore). Each tick gets its own sync.Map; old maps
	// are GC'd once no goroutine holds a reference.
	// Per spec R7: "MUST be emitted at most ONCE per evaluation cycle per target."
	dependencyWaitLogged atomic.Pointer[sync.Map]
}

// dispatcher is the package-level singleton.
var disp = &dispatcherState{}

// trackGoroutineLocked is the unlocked variant of trackGoroutine. The caller
// MUST already hold d.mu. It checks the shutdown flag and increments the
// WaitGroup counter atomically with respect to Stop(). Returns true if a new
// goroutine can be safely launched (caller MUST then call d.wg.Done() when it
// exits). Returns false if shutdown is in progress; caller must NOT spawn.
//
// Used by Start() to keep the entire start sequence (state setup +
// goroutine-launch decision) under a single critical section, eliminating the
// race window where Stop() could land between unlock and trackGoroutine and
// leave the dispatcher in a half-started state (started=true, no worker).
func (d *dispatcherState) trackGoroutineLocked() bool {
	if d.shuttingDown.Load() {
		return false
	}
	d.wg.Add(1)
	return true
}

// trackGoroutine atomically checks the shutdown flag and increments the
// WaitGroup counter under disp.mu. Returns true if a new goroutine can be
// safely launched (caller MUST then call disp.wg.Done() when it exits).
// Returns false if shutdown is in progress; caller must NOT spawn the
// goroutine.
//
// This avoids a TOCTOU race where Stop() could observe wg counter at zero
// and unblock Wait() between a check of shuttingDown and the wg.Add(1) call,
// triggering a "WaitGroup is reused before previous Wait has returned" panic.
// The mutex in Stop() (acquired BEFORE setting shuttingDown=true) fences any
// in-flight trackGoroutine call.
func (d *dispatcherState) trackGoroutine() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.trackGoroutineLocked()
}

// Start launches the 60s safety-net tick and prepares semaphores. Idempotent —
// only the first call starts the goroutine. If provider is nil, dispatch will
// skip sends (Phase 2B: provider wired in Phase 3).
func Start(parentCtx context.Context, db *gorm.DB, provider StreamProvider) {
	if !disp.started.CompareAndSwap(false, true) {
		return // already running
	}

	disp.mu.Lock()
	disp.ctx, disp.cancel = context.WithCancel(parentCtx)
	disp.db = db
	disp.provider = provider
	disp.globalSem = semaphore.NewWeighted(int64(globalConcurrency()))
	disp.ticker = time.NewTicker(60 * time.Second)
	disp.dependencyWaitLogged.Store(&sync.Map{})
	ctx := disp.ctx
	ticker := disp.ticker

	// Decide whether to spawn the safety-net worker WHILE STILL HOLDING disp.mu.
	// This closes the race where Stop() could land between releasing the lock
	// and calling trackGoroutine, leaving started=true with no worker (a stuck
	// half-started state from which subsequent Start calls would silently no-op).
	if !disp.trackGoroutineLocked() {
		// Shutdown happened (or was already in progress). Roll back started so a
		// future Start call after a fresh process boot — or more importantly,
		// after Stop completes — can re-initialize cleanly. Note: the ticker
		// and context we just created here will be cleaned up by Stop's own
		// ticker.Stop() / cancel(), so no leak.
		disp.started.Store(false)
		disp.mu.Unlock()
		return
	}
	disp.mu.Unlock()

	go func() {
		defer disp.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runSafetyNetTick(ctx, db)
			}
		}
	}()
}

// Stop cancels the dispatcher context but does NOT wait. Call WaitDone to block.
// Acquires disp.mu BEFORE setting shuttingDown so any in-flight call to
// trackGoroutine fences cleanly: either it ran and the WaitGroup counter is
// already incremented (Stop will see it via Wait), or it will run after Stop
// and observe shuttingDown=true (and refuse to Add).
func Stop() {
	disp.mu.Lock()
	disp.shuttingDown.Store(true)
	cancel := disp.cancel
	ticker := disp.ticker
	disp.mu.Unlock()
	if ticker != nil {
		ticker.Stop()
	}
	if cancel != nil {
		cancel()
	}
}

// WaitDone blocks until all in-flight dispatches return, or until timeout elapses.
// Returns true if drained cleanly, false on timeout.
func WaitDone(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		disp.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// IsShuttingDown reports whether Stop has been called.
func IsShuttingDown() bool {
	return disp.shuttingDown.Load()
}

// dispatcherOnAgentConnect is the internal entry point called by hooks.go
// when an agent's stream registers. Spawns a goroutine that queries
// PendingForAgent and dispatches each target.
// No-op if shutting down. The public hook OnAgentConnect in hooks.go owns
// the kill-switch check and delegates here.
func dispatcherOnAgentConnect(agentID uint) {
	if IsShuttingDown() {
		return
	}

	disp.mu.Lock()
	ctx := disp.ctx
	db := disp.db
	disp.mu.Unlock()

	if ctx == nil || db == nil {
		return
	}

	if !disp.trackGoroutine() {
		return
	}
	go func() {
		defer disp.wg.Done()

		pending, err := PendingForAgent(db, agentID)
		if err != nil {
			return
		}
		for _, t := range pending {
			if IsShuttingDown() {
				return
			}
			dispatch(ctx, db, t)
		}
	}()
}

// dispatcherOnCommandResult is the internal entry point called by hooks.go
// when a CommandResult arrives. If the cmd_id matches a pending recovery
// dispatch, handles success/failure and returns true. If not, returns false
// (caller handles it normally). The public hook OnCommandResult in hooks.go
// owns the nil-check and delegates here.
func dispatcherOnCommandResult(result *agentpb.CommandResult) bool {
	if result == nil {
		return false
	}
	cmdID := result.GetCmdId()
	if cmdID == "" {
		return false
	}

	// Look up cmd_id in pendingResults.
	val, ok := disp.pendingResults.LoadAndDelete(cmdID)
	if !ok {
		return false
	}
	targetID, ok := val.(uint)
	if !ok {
		return false
	}

	disp.mu.Lock()
	db := disp.db
	disp.mu.Unlock()

	if db == nil {
		return true // we own this cmd_id even if we can't process
	}

	// Find the target to know its recovery (for success_pattern + logging).
	target, err := findTargetByID(db, targetID)
	if err != nil {
		return true
	}
	rec, err := findRecoveryByID(db, target.RecoveryID)
	if err != nil {
		return true
	}

	output := result.GetResult()
	successPattern := rec.SuccessPattern
	if successPattern == "" {
		successPattern = DefaultSuccessPattern
	}

	if matchesSuccessPattern(output, successPattern) {
		_ = SetTargetSucceeded(db, targetID, output)
		LogRecoverySucceeded(rec.YamlID, target.AgentID, cmdID)
	} else {
		terminal := target.Attempts >= rec.MaxAttempts
		_ = SetTargetFailed(db, targetID, truncateStr(output, 512), terminal)
		LogRecoveryFailed(rec.YamlID, target.AgentID, cmdID, truncateStr(output, 512), terminal)
	}
	return true
}

// matchesSuccessPattern returns true if any complete line of output equals the
// sentinel pattern (after trimming surrounding whitespace and a trailing CR).
// This is stricter than substring matching: a script that prints
// "__UTMSTACK_RECOVERY_OK__ NOT FOUND" inside an error message will NOT pass.
// Recovery scripts are contractually required to emit the sentinel as its own
// final stdout line on success.
func matchesSuccessPattern(output, pattern string) bool {
	if pattern == "" {
		return false
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == pattern {
			return true
		}
	}
	return false
}

// runSafetyNetTick executes the 60s safety-net sweep in order.
func runSafetyNetTick(ctx context.Context, db *gorm.DB) {
	if IsShuttingDown() {
		return
	}

	// Reset per-tick dedup map by atomically swapping in a fresh sync.Map.
	// Concurrent dispatches use whichever pointer they observed at LoadOrStore
	// time; old maps are GC'd once their last reference is released.
	disp.dependencyWaitLogged.Store(&sync.Map{})

	now := time.Now().UTC()

	// Step 1: Atomically mark expired recoveries AND their non-terminal
	// targets in a single transaction. The returned yaml_ids are exactly the
	// recoveries that transitioned this tick (source of truth for logging).
	// Spec R8: "Non-terminal targets of an expired recovery MUST transition to EXPIRED."
	if expiredIDs, err := ExpireRecoveriesAndTargets(db, now); err == nil {
		for _, yamlID := range expiredIDs {
			LogRecoveryExpired(yamlID)
		}
	}

	// Step 2: Re-arm DISPATCHED targets past ack_timeout.
	// Spec R3 MUST: "re-armed identically with last_error = \"ack_timeout\"".
	// Direct re-arm to PENDING (not via FAILED) to match spec — avoids the
	// retry_after delay that SetTargetFailed + EligibleForRetry would introduce.
	// Only targets that have exhausted attempts go to terminal FAILED.
	timedOut, err := DispatchedPastTimeout(db, now)
	if err == nil {
		for _, tw := range timedOut {
			if tw.Target.Attempts >= tw.Recovery.MaxAttempts {
				// Max attempts exhausted — terminal FAILED.
				_ = SetTargetFailed(db, tw.Target.ID, "ack_timeout", true)
				LogRecoveryFailed(tw.Recovery.YamlID, tw.Target.AgentID, tw.Target.LastCmdID, "ack_timeout", true)
			} else {
				// Re-arm directly to PENDING with last_error="ack_timeout".
				// SetTargetStatus(PENDING) clears completed_at but not last_error;
				// update last_error explicitly afterwards.
				_ = SetTargetStatus(db, tw.Target.ID, TargetStatusPending)
				_ = db.Model(&models.RecoveryTarget{}).
					Where("id = ?", tw.Target.ID).
					Update("last_error", "ack_timeout")
				retryAt := now // immediate re-arm, no delay
				LogRecoveryRetryScheduled(tw.Recovery.YamlID, tw.Target.AgentID, retryAt, tw.Target.Attempts)
			}
		}
	}

	// Step 3: Re-arm eligible-for-retry targets.
	retryEligible, err := EligibleForRetry(db, now)
	if err == nil {
		for _, tw := range retryEligible {
			_ = SetTargetStatus(db, tw.Target.ID, TargetStatusPending)
			retryAt := now.Add(time.Duration(tw.Recovery.RetryAfterSecs) * time.Second)
			LogRecoveryRetryScheduled(tw.Recovery.YamlID, tw.Target.AgentID, retryAt, tw.Target.Attempts)
		}
	}

	// Step 4: Online-agents-with-pending sweep.
	// Fetch all ACTIVE recoveries' PENDING targets and dispatch for online agents.
	sweepDispatch(ctx, db, now)

	// Step 5: Recompute COMPLETED status for fully-terminal recoveries.
	recomputeCompleted(db)
}

// sweepDispatch finds agents with pending targets that are online, dispatches them.
func sweepDispatch(ctx context.Context, db *gorm.DB, _ time.Time) {
	if IsShuttingDown() {
		return
	}

	provider := getProvider()
	if provider == nil {
		return
	}

	// Get all ACTIVE recoveries, fetch pending targets for connected agents.
	activeRecoveries, err := AllByStatus(db, RecoveryStatusActive)
	if err != nil {
		return
	}

	for _, r := range activeRecoveries {
		if IsShuttingDown() {
			return
		}
		// Get all pending targets for this recovery.
		targets, tErr := ListTargets(db, r.ID)
		if tErr != nil {
			continue
		}
		for _, t := range targets {
			if TargetStatus(t.Status) != TargetStatusPending {
				continue
			}
			if !provider.IsOnline(t.AgentID) {
				continue
			}
			tw := TargetWithRecovery{Target: t, Recovery: r}
			dispatch(ctx, db, tw)
		}
	}
}

// recomputeCompleted marks an ACTIVE recovery as COMPLETED only when every
// target is in a TRULY terminal state. A FAILED target with attempts <
// max_attempts is NOT terminal — it is awaiting EligibleForRetry to re-arm it
// to PENDING. Treating it as terminal here causes the recovery to be marked
// COMPLETED prematurely; sweepDispatch then orphans the re-armed PENDING target
// because it only iterates ACTIVE recoveries (E2E TEST 2 regression).
//
// Strategy (Option A): instead of inspecting aggregate counts, query directly
// for any non-terminal target. A target is non-terminal if:
//   - status IN (PENDING, DISPATCHED) — always non-terminal, OR
//   - status = FAILED AND attempts < max_attempts — retryable, not terminal.
//
// SUCCEEDED, EXPIRED, SKIPPED are always terminal.
// BLOCKED is terminal for completion purposes (operator resets via CLI retry).
// FAILED with attempts >= max_attempts is terminal.
func recomputeCompleted(db *gorm.DB) {
	activeRecoveries, err := AllByStatus(db, RecoveryStatusActive)
	if err != nil {
		return
	}
	for _, r := range activeRecoveries {
		// Require at least one target to exist before declaring completion.
		var totalCount int64
		if err := db.Model(&models.RecoveryTarget{}).
			Where("recovery_id = ? AND deleted_at IS NULL", r.ID).
			Count(&totalCount).Error; err != nil {
			continue
		}
		if totalCount == 0 {
			continue
		}

		// Count targets that are NOT yet truly terminal.
		// PENDING or DISPATCHED: always non-terminal.
		// FAILED with attempts < max_attempts: retryable, non-terminal.
		var nonTerminalCount int64
		err := db.Model(&models.RecoveryTarget{}).
			Where("recovery_id = ? AND deleted_at IS NULL", r.ID).
			Where(
				"status IN ? OR (status = ? AND attempts < ?)",
				[]string{
					string(TargetStatusPending),
					string(TargetStatusDispatched),
				},
				string(TargetStatusFailed),
				r.MaxAttempts,
			).
			Count(&nonTerminalCount).Error
		if err != nil {
			continue
		}

		if nonTerminalCount == 0 {
			_ = SetStatus(db, r.ID, RecoveryStatusCompleted, "")
		}
	}
}

// dispatch is the core dispatch function. It handles dependency checks,
// semaphore acquisition, dry-run, stream send, and state updates.
func dispatch(ctx context.Context, db *gorm.DB, tw TargetWithRecovery) {
	if IsShuttingDown() {
		return
	}

	r := &tw.Recovery
	t := &tw.Target

	// Dependency check: if recovery has depends_on, verify the dep is satisfied.
	if r.DependsOnYamlID != "" {
		depRec, depErr := findRecoveryByYAMLID(db, r.DependsOnYamlID)
		if depErr != nil {
			// Dep recovery missing entirely → block this target (terminal).
			// Spec R7: "recovery_blocked MUST be logged" on terminal dep failure.
			_ = SetTargetStatus(db, t.ID, TargetStatusBlocked)
			LogRecoveryBlocked(r.YamlID, "target_dep_terminal_failed",
				r.DependsOnYamlID+":missing")
			return
		}

		depStatus := RecoveryStatus(depRec.Status)
		switch depStatus {
		case RecoveryStatusBlocked, RecoveryStatusArchived, RecoveryStatusExpired:
			// Dep recovery is in a terminal-failure state → block this target.
			// Spec R7: "recovery_blocked MUST be logged" on terminal dep failure.
			_ = SetTargetStatus(db, t.ID, TargetStatusBlocked)
			LogRecoveryBlocked(r.YamlID, "target_dep_terminal_failed",
				r.DependsOnYamlID+":"+string(depStatus))
			return
		default:
			// Check if dep target for this specific agent is SUCCEEDED.
			depTarget, tErr := FindTarget(db, depRec.ID, t.AgentID)
			if tErr != nil {
				// Dep target doesn't exist for this agent yet — wait (transient).
				logDependencyWaitOnce(t.ID, r.YamlID, t.AgentID, r.DependsOnYamlID)
				return
			}
			switch TargetStatus(depTarget.Status) {
			case TargetStatusSucceeded:
				// dep satisfied, proceed
			case TargetStatusFailed, TargetStatusExpired, TargetStatusSkipped:
				// Dep target reached terminal failure → block this target (terminal).
				// Spec R7: "recovery_blocked MUST be logged" on terminal dep failure.
				_ = SetTargetStatus(db, t.ID, TargetStatusBlocked)
				LogRecoveryBlocked(r.YamlID, "target_dep_terminal_failed",
					r.DependsOnYamlID+":"+depTarget.Status)
				return
			default:
				// Dep still pending/dispatched/blocked — wait (transient).
				logDependencyWaitOnce(t.ID, r.YamlID, t.AgentID, r.DependsOnYamlID)
				return
			}
		}
	}

	// Acquire global semaphore — context-aware so shutdown cancels.
	if err := disp.globalSem.Acquire(ctx, 1); err != nil {
		// Context cancelled (shutdown). Do not mark as failed.
		return
	}
	defer disp.globalSem.Release(1)

	// Acquire per-recovery semaphore.
	perSem := getOrCreatePerRecoverySem(r.ID, int64(r.MaxConcurrency))
	if err := perSem.Acquire(ctx, 1); err != nil {
		return
	}
	defer perSem.Release(1)

	// Dry-run branch: log + mark succeeded without sending.
	// No cmd_id is generated — there is no real send to correlate.
	if r.DryRun {
		_ = SetTargetSucceeded(db, t.ID, "[DRY-RUN] would have dispatched")
		LogRecoveryDryRun(r.YamlID, t.AgentID)
		return
	}

	// Get stream from provider.
	provider := getProvider()
	if provider == nil {
		_ = SetTargetFailed(db, t.ID, "dispatch disabled: no stream provider", false)
		return
	}

	stream, ok := provider.GetStream(t.AgentID)
	if !ok {
		_ = SetTargetFailed(db, t.ID, "agent offline", false)
		return
	}

	// Generate cmd_id.
	cmdID := uuid.New().String()

	// Mark target as DISPATCHED FIRST via a guarded UPDATE (status=PENDING).
	// This atomically claims the target so concurrent workers (e.g. the
	// safety-net sweep tick and an OnAgentConnect goroutine) cannot both
	// dispatch the same target. If rowsAffected=0, another worker won.
	rowsAffected, err := SetTargetDispatched(db, t.ID, cmdID)
	if err != nil {
		_ = SetTargetFailed(db, t.ID, fmt.Sprintf("db error before send: %v", err), false)
		return
	}
	if rowsAffected == 0 {
		// Another worker already claimed this target — bail out cleanly
		// (defer'd semaphore releases will run).
		return
	}

	// Now that we own the dispatch, record cmd_id → target_id for ACK matching.
	disp.pendingResults.Store(cmdID, t.ID)

	// Build the UtmCommand proto message.
	cmd := &agentpb.BidirectionalStream{
		StreamMessage: &agentpb.BidirectionalStream_Command{
			Command: &agentpb.UtmCommand{
				AgentId: fmt.Sprintf("%d", t.AgentID),
				Command: r.Script,
				CmdId:   cmdID,
				Shell:   r.Shell,
			},
		},
	}

	// Acquire per-agent stream mutex to serialize with panel-driven sends.
	// Use defer for unlock so a panic in stream.Send doesn't leak the mutex
	// (gRPC streams can panic on closed/aborted connections).
	mu := StreamMutex.For(t.AgentID)
	var sendErr error
	func() {
		mu.Lock()
		defer mu.Unlock()
		sendErr = stream.Send(cmd)
	}()

	if sendErr != nil {
		disp.pendingResults.Delete(cmdID)
		_ = SetTargetFailed(db, t.ID, fmt.Sprintf("send_error: %v", sendErr), false)
		return
	}

	LogRecoveryDispatched(r.YamlID, t.AgentID, cmdID, t.Attempts+1)
	// pendingResults entry remains until OnCommandResult processes the ACK.
}

// getOrCreatePerRecoverySem lazily creates or returns an existing per-recovery semaphore.
func getOrCreatePerRecoverySem(recoveryID uint, weight int64) *semaphore.Weighted {
	if weight <= 0 {
		weight = 100
	}
	actual, _ := disp.perRecoverySem.LoadOrStore(recoveryID, semaphore.NewWeighted(weight))
	return actual.(*semaphore.Weighted)
}

// getProvider returns the current StreamProvider in a thread-safe manner.
func getProvider() StreamProvider {
	disp.mu.Lock()
	p := disp.provider
	disp.mu.Unlock()
	return p
}

// findRecoveryByID is a thin DB helper used only by this file.
func findRecoveryByID(db *gorm.DB, id uint) (*models.Recovery, error) {
	var r models.Recovery
	if err := db.Where("id = ?", id).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// findRecoveryByYAMLID is a thin DB helper used only by this file.
func findRecoveryByYAMLID(db *gorm.DB, yamlID string) (*models.Recovery, error) {
	var r models.Recovery
	if err := db.Where("yaml_id = ?", yamlID).First(&r).Error; err != nil {
		return nil, fmt.Errorf("find recovery yaml_id=%q: %w", yamlID, err)
	}
	return &r, nil
}

// findTargetByID is a thin DB helper used only by this file.
func findTargetByID(db *gorm.DB, id uint) (*models.RecoveryTarget, error) {
	var t models.RecoveryTarget
	if err := db.Where("id = ?", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// truncateStr truncates s to at most n bytes.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// logDependencyWaitOnce emits LogRecoveryDependencyWait at most once per
// evaluation cycle per target. Loads the current dedup-map pointer atomically;
// the pointer is swapped to a fresh map at the start of each runSafetyNetTick.
func logDependencyWaitOnce(targetID uint, yamlID string, agentID uint, blockedBy string) {
	m := disp.dependencyWaitLogged.Load()
	if m == nil {
		// Dispatcher not started yet (defensive). Just log without dedup.
		LogRecoveryDependencyWait(yamlID, agentID, blockedBy)
		return
	}
	if _, alreadyLogged := m.LoadOrStore(targetID, struct{}{}); !alreadyLogged {
		LogRecoveryDependencyWait(yamlID, agentID, blockedBy)
	}
}

// globalConcurrency returns the configured global concurrency cap.
func globalConcurrency() int {
	return configGlobalConcurrency
}
