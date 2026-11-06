package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
)

func init() { gin.SetMode(gin.TestMode) }

// SOAR-F4b: a plain access token (what every authenticated caller already
// has, just to reach this handler) must not pass as the console's "prove you
// still know the password" second factor.
func TestRequirePasswordTokenRejectsAPlainAccessToken(t *testing.T) {
	signer := jwtpkg.NewSigner("test-secret", time.Hour)
	userID := uuid.New()
	h := &CommandWSHandler{signer: signer}

	accessToken, _, err := signer.Sign(userID, "user@x.com", nil, nil, uuid.New(), "tenant-a", false)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x?password_token="+accessToken, nil)

	if h.requirePasswordToken(c, &authz.Actor{UserID: userID}) {
		t.Fatal("a plain access token must not satisfy the password-proof check")
	}
}

// A genuine password-proof token for a DIFFERENT user must not satisfy the
// check either — it has to be tied to the specific caller.
func TestRequirePasswordTokenRejectsAnotherUsersProof(t *testing.T) {
	signer := jwtpkg.NewSigner("test-secret", time.Hour)
	otherUser := uuid.New()
	callingUser := uuid.New()
	h := &CommandWSHandler{signer: signer}

	proof, err := signer.SignEmpty(otherUser, 5*time.Minute)
	if err != nil {
		t.Fatalf("SignEmpty: %v", err)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x?password_token="+proof, nil)

	if h.requirePasswordToken(c, &authz.Actor{UserID: callingUser}) {
		t.Fatal("another user's password-proof token must not satisfy the check")
	}
}

// The real, correctly-bound case must still work.
func TestRequirePasswordTokenAcceptsOwnProof(t *testing.T) {
	signer := jwtpkg.NewSigner("test-secret", time.Hour)
	userID := uuid.New()
	h := &CommandWSHandler{signer: signer}

	proof, err := signer.SignEmpty(userID, 5*time.Minute)
	if err != nil {
		t.Fatalf("SignEmpty: %v", err)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x?password_token="+proof, nil)

	if !h.requirePasswordToken(c, &authz.Actor{UserID: userID}) {
		t.Fatal("the caller's own genuine password-proof token should be accepted")
	}
}
