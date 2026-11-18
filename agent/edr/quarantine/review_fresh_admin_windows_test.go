//go:build windows && edr_acceptance

package quarantine

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"golang.org/x/sys/windows"
)

type freshAdminCase struct {
	Root       string                 `json:"root"`
	Database   string                 `json:"database"`
	ID         string                 `json:"id"`
	Operation  string                 `json:"operation"`
	Inspection FreshRestoreInspection `json:"inspection"`
}

func TestFreshElevatedAdministratorWorker(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_ADMIN_WORKER") != "1" {
		t.Skip("owned elevated-administrator fresh worker")
	}
	if !windows.GetCurrentProcessToken().IsElevated() {
		t.Fatal("owned administrator worker is not elevated")
	}
	body, err := os.ReadFile(os.Getenv("OPENEDR_FRESH_ADMIN_CASE"))
	if err != nil {
		t.Fatal(err)
	}
	var input freshAdminCase
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	a, err := ApproveFreshRestore(input.Root, input.ID, input.Operation, input.Inspection, "administrator independently approves fresh restore")
	if err != nil {
		t.Fatal(err)
	}
	body, err = json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	reached := false
	_, err = executeFreshRestore(input.Root, body, "administrator stages reviewed fresh restore", freshTestCacheOpener(input.Database), func(at string) error {
		if at == "fresh-execution-verified" {
			reached = true
			return errors.New("owned administrator verified-stage checkpoint")
		}
		return nil
	})
	if err == nil || !reached {
		t.Fatal("administrator could not prepare a verified fresh stage", err)
	}
	stage, err := openAdmissionRoot(filepath.Join(string(input.Inspection.Destination.ParentPath), ".edr-restore-"+input.Operation), false)
	if err != nil {
		t.Fatal(err)
	}
	stageID, err := admissionFileID(stage.file)
	if err != nil {
		stage.Close()
		t.Fatal(err)
	}
	sd, err := windows.GetSecurityInfo(windows.Handle(stage.file.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	stage.Close()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("administrator_actor_sid=%s stage_id=%s stage_security=%s", a.ActorID, stageID, sd.String())
	t.Log("owned elevated administrator retained a verified fresh stage")
}

func freshAdminLogon(t *testing.T, user, password string) windows.Token {
	t.Helper()
	name, err := windows.UTF16PtrFromString(user)
	if err != nil {
		t.Fatal(err)
	}
	domain, _ := windows.UTF16PtrFromString(".")
	secret, err := windows.UTF16PtrFromString(password)
	if err != nil {
		t.Fatal(err)
	}
	var token windows.Token
	proc := windows.NewLazySystemDLL("advapi32.dll").NewProc("LogonUserW")
	ok, _, callErr := proc.Call(uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(domain)), uintptr(unsafe.Pointer(secret)), 2, 0, uintptr(unsafe.Pointer(&token)))
	if ok == 0 {
		t.Fatal("owned local administrator logon failed", callErr)
	}
	if token.IsElevated() {
		return token
	}
	linked, err := token.GetLinkedToken()
	token.Close()
	if err != nil {
		t.Fatal("owned administrator linked token unavailable", err)
	}
	if !linked.IsElevated() {
		linked.Close()
		t.Fatal("owned administrator token is not elevated")
	}
	return linked
}

// LogonUser creates a token without the first interactive logon's profile.
// Create only this disposable account's profile so os/user.Current can obtain
// its home directory; never substitute an actor SID or weaken production checks.
func freshAdminProfile(t *testing.T, token windows.Token, user string) {
	t.Helper()
	u, err := token.GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sid, err := windows.UTF16PtrFromString(u.User.Sid.String())
	if err != nil {
		t.Fatal(err)
	}
	name, err := windows.UTF16PtrFromString(user)
	if err != nil {
		t.Fatal(err)
	}
	buffer := make([]uint16, windows.MAX_PATH)
	userenv := windows.NewLazySystemDLL("userenv.dll")
	status, _, _ := userenv.NewProc("CreateProfile").Call(uintptr(unsafe.Pointer(sid)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if int32(status) < 0 {
		t.Fatalf("owned local profile creation failed: HRESULT %#x", uint32(status))
	}
	if _, err = token.GetUserProfileDirectory(); err != nil {
		t.Fatal("owned local profile cannot be queried", err)
	}
	t.Cleanup(func() {
		ok, _, e := userenv.NewProc("DeleteProfileW").Call(uintptr(unsafe.Pointer(sid)), 0, 0)
		if ok == 0 {
			t.Error("owned local profile cleanup failed", e)
		}
	})
}

func TestFreshAdministratorStageRemainsProvableToSystem(t *testing.T) {
	if os.Getenv("OPENEDR_FRESH_ADMIN_INTEROP") != "1" {
		t.Skip("requires an owned disposable SYSTEM Windows lab")
	}
	current, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil || !current.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
		t.Fatal("native interoperability control must run as SYSTEM", err)
	}
	// Credentials exist only for this new local account in the disposable VM;
	// never log them, place them in manifests, or use an existing account.
	user := "edrf" + uuid.NewString()[:8]
	random := make([]byte, 24)
	if _, err = rand.Read(random); err != nil {
		t.Fatal(err)
	}
	password := base64.RawURLEncoding.EncodeToString(random) + "aA1!"
	net := func(args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, filepath.Join(os.Getenv("SystemRoot"), "System32", "net.exe"), args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("owned account setup failed: %w: %s", err, strings.ReplaceAll(string(output), password, "[REDACTED]"))
		}
		return nil
	}
	if err = net("user", user, password, "/add", "/y"); err != nil {
		t.Fatal("could not create owned local test account", err)
	}
	defer func() {
		if e := net("user", user, "/delete"); e != nil {
			t.Error("could not remove owned local test account", e)
		}
	}()
	if err = net("localgroup", "Administrators", user, "/add"); err != nil {
		t.Fatal("could not grant owned test administrator role", err)
	}
	token := freshAdminLogon(t, user, password)
	defer token.Close()
	freshAdminProfile(t, token, user)
	s, id, destination, stored, a, _ := freshExecutionFixture(t)
	input := freshAdminCase{Root: s.dir, Database: filepath.Join(filepath.Dir(s.dir), "edr.db"), ID: id, Operation: uuid.NewString(), Inspection: a.Inspection}
	control := filepath.Join(filepath.Dir(s.dir), "owned-administrator-case.json")
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(control, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.cache.Close(); err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, self, "-test.run=^TestFreshElevatedAdministratorWorker$", "-test.count=1", "-test.v")
	cmd.SysProcAttr = &syscall.SysProcAttr{Token: syscall.Token(token), HideWindow: true}
	cmd.Env = append(os.Environ(), "OPENEDR_FRESH_ADMIN_WORKER=1", "OPENEDR_FRESH_ADMIN_CASE="+control)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("owned administrator process failed: %v %s", err, output)
	}
	t.Logf("distinct administrator process evidence: %s", output)
	root, err := s.openRoot()
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	queue, item, op, err := openFreshApproval(root, id, input.Operation)
	if err != nil {
		t.Fatal("SYSTEM cannot reopen independently approved administrator metadata", err)
	}
	defer queue.Close()
	defer item.Close()
	defer op.Close()
	approved, err := readFreshApproval(root, queue, item, op, id, input.Operation)
	if err != nil {
		t.Fatal("SYSTEM cannot verify administrator approval", err)
	}
	if approved.ActorID == a.ActorID {
		t.Fatal("administrator decision was falsely attributed to SYSTEM")
	}
	stagePath := filepath.Join(filepath.Dir(destination), ".edr-restore-"+input.Operation)
	stage, err := openAdmissionRoot(stagePath, false)
	if err != nil {
		t.Fatal("SYSTEM cannot reopen administrator private stage", err)
	}
	stageID, err := admissionFileID(stage.file)
	if err != nil {
		stage.Close()
		t.Fatal(err)
	}
	proof := managedFreshRestoreStage(stage, stageID)
	stage.Close()
	if !errors.Is(proof, ErrAlreadyQuarantined) {
		t.Fatal("SYSTEM lost administrator-stage ownership proof", proof)
	}
	requirePurgePayload(t, filepath.Join(stagePath, "snapshot"), "payload")
	requirePurgePayload(t, stored, "payload")
	t.Log("distinct administrator actor and SYSTEM reader retain the same private stage")
}
