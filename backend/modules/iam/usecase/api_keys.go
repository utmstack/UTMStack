package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type apiKeyUsecase struct {
	repo     connectors.APIKeyRepository
	userRepo connectors.UserRepository
}

func NewAPIKeyUsecase(repo connectors.APIKeyRepository, userRepo connectors.UserRepository) connectors.APIKeyUsecase {
	return &apiKeyUsecase{repo: repo, userRepo: userRepo}
}

func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func joinIPs(ips []string) string {
	cleaned := make([]string, 0, len(ips))
	for _, ip := range ips {
		if s := strings.TrimSpace(ip); s != "" {
			cleaned = append(cleaned, s)
		}
	}
	return strings.Join(cleaned, ",")
}

func splitIPs(csv string) []string {
	var out []string
	for _, ip := range strings.Split(csv, ",") {
		if s := strings.TrimSpace(ip); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func toAPIKeyResponse(k *domain.APIKey) *dto.APIKeyResponse {
	return &dto.APIKeyResponse{
		ID:          k.ID,
		Name:        k.Name,
		AllowedIP:   splitIPs(k.AllowedIP),
		CreatedAt:   k.CreatedAt,
		GeneratedAt: k.GeneratedAt,
		ExpiresAt:   k.ExpiresAt,
	}
}

func (u *apiKeyUsecase) Create(ctx context.Context, userID uint64, req dto.APIKeyUpsertRequest) (*dto.APIKeyResponse, error) {
	dup, err := u.repo.FindByNameAndUser(ctx, req.Name, userID)
	if err != nil {
		return nil, err
	}
	if dup != nil {
		return nil, domain.ErrAPIKeyNameTaken
	}
	key, err := generateKey()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	k := &domain.APIKey{
		UserID:      userID,
		Name:        req.Name,
		APIKey:      key,
		AllowedIP:   joinIPs(req.AllowedIP),
		CreatedAt:   now,
		GeneratedAt: now,
		ExpiresAt:   req.ExpiresAt,
	}
	if err := u.repo.Create(ctx, k); err != nil {
		return nil, err
	}
	return toAPIKeyResponse(k), nil
}

func (u *apiKeyUsecase) Update(ctx context.Context, userID, id uint64, req dto.APIKeyUpsertRequest) (*dto.APIKeyResponse, error) {
	k, err := u.repo.FindByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, domain.ErrAPIKeyNotFound
	}
	if req.Name != k.Name {
		dup, err := u.repo.FindByNameAndUser(ctx, req.Name, userID)
		if err != nil {
			return nil, err
		}
		if dup != nil {
			return nil, domain.ErrAPIKeyNameTaken
		}
	}
	k.Name = req.Name
	k.AllowedIP = joinIPs(req.AllowedIP)
	k.ExpiresAt = req.ExpiresAt
	if err := u.repo.Update(ctx, k); err != nil {
		return nil, err
	}
	return toAPIKeyResponse(k), nil
}

func (u *apiKeyUsecase) Delete(ctx context.Context, userID, id uint64) error {
	return u.repo.DeleteByIDAndUser(ctx, id, userID)
}

func (u *apiKeyUsecase) Get(ctx context.Context, userID, id uint64) (*dto.APIKeyResponse, error) {
	k, err := u.repo.FindByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, domain.ErrAPIKeyNotFound
	}
	return toAPIKeyResponse(k), nil
}

func (u *apiKeyUsecase) List(ctx context.Context, userID uint64, q dto.ListAPIKeysQuery) (*dto.APIKeyListResponse, error) {
	keys, total, err := u.repo.ListByUser(ctx, userID, q.Page, q.PageSize)
	if err != nil {
		return nil, err
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	out := make([]dto.APIKeyResponse, 0, len(keys))
	for i := range keys {
		out = append(out, *toAPIKeyResponse(&keys[i]))
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &dto.APIKeyListResponse{
		Data: out,
		PageInfo: dto.PageInfo{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (u *apiKeyUsecase) Generate(ctx context.Context, userID, id uint64) (string, error) {
	k, err := u.repo.FindByIDAndUser(ctx, id, userID)
	if err != nil {
		return "", err
	}
	if k == nil {
		return "", domain.ErrAPIKeyNotFound
	}
	key, err := generateKey()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	if k.ExpiresAt != nil {
		if dur := k.ExpiresAt.Sub(k.GeneratedAt); dur > 0 {
			exp := now.Add(dur)
			k.ExpiresAt = &exp
		}
	}
	k.APIKey = key
	k.GeneratedAt = now
	if err := u.repo.Update(ctx, k); err != nil {
		return "", err
	}
	return key, nil
}

func (u *apiKeyUsecase) Authenticate(ctx context.Context, apiKey, clientIP string) (*connectors.APIKeyAuthResult, error) {
	k, err := u.repo.FindByKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if k == nil || !ipAllowed(k.AllowedIP, clientIP) || k.Expired(time.Now().UTC()) {
		return nil, domain.ErrAPIKeyInvalid
	}
	user, err := u.userRepo.FindByID(ctx, k.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Activated {
		return nil, domain.ErrAPIKeyInvalid
	}
	perms, err := u.userRepo.FindPermissionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	roleRows, err := u.userRepo.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(roleRows))
	for _, r := range roleRows {
		roles = append(roles, r.Name)
	}
	return &connectors.APIKeyAuthResult{
		UserID:      user.ID,
		Login:       user.Login,
		Email:       user.Email,
		Roles:       roles,
		Permissions: perms,
	}, nil
}

// ipAllowed reports whether remoteIP is permitted by the CSV allowlist (IPs and
// CIDR ranges). An empty allowlist permits everything.
func ipAllowed(allowedCSV, remoteIP string) bool {
	if strings.TrimSpace(allowedCSV) == "" {
		return true
	}
	ip := net.ParseIP(remoteIP)
	for _, entry := range splitIPs(allowedCSV) {
		if strings.Contains(entry, "/") {
			if _, ipnet, err := net.ParseCIDR(entry); err == nil && ip != nil && ipnet.Contains(ip) {
				return true
			}
			continue
		}
		if entry == remoteIP {
			return true
		}
	}
	return false
}
