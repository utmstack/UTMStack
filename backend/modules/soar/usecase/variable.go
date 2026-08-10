package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

var variableInterpolationRegex = regexp.MustCompile(`\$\[variables\.([^\]]+)\]`)

const maskedValue = "****"

type variableUsecase struct {
	repo   connectors.VariableRepository
	cipher connectors.VariableCipher
}

func NewVariableUsecase(repo connectors.VariableRepository, cipher connectors.VariableCipher) connectors.VariableUsecase {
	return &variableUsecase{repo: repo, cipher: cipher}
}

func (u *variableUsecase) Create(ctx context.Context, req dto.CreateVariableRequest, user string) (*dto.VariableResponse, error) {
	now := time.Now().UTC()
	v := &domain.SoarVariable{
		Name:        req.Name,
		Description: req.Description,
		IsSecret:    req.IsSecret,
		CreatedBy:   user,
		CreatedAt:   now,
		ModifiedBy:  user,
		ModifiedAt:  &now,
	}

	value := req.Value
	if req.IsSecret {
		encrypted, err := u.cipher.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("variableUsecase.Create: encrypt: %w", err)
		}
		value = encrypted
	}
	v.Value = value

	if err := u.repo.Save(ctx, v); err != nil {
		return nil, fmt.Errorf("variableUsecase.Create: %w", err)
	}
	return u.toResponse(v), nil
}

func (u *variableUsecase) Update(ctx context.Context, req dto.UpdateVariableRequest, user string) (*dto.VariableResponse, error) {
	v, err := u.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil && *req.Name != "" {
		v.Name = *req.Name
	}
	if req.Description != nil {
		v.Description = req.Description
	}

	plain, err := u.plainValue(v, req)
	if err != nil {
		return nil, err
	}
	if plain == "" {
		return nil, domain.ErrVariableValueRequired
	}

	v.IsSecret = req.IsSecret
	if req.IsSecret {
		encrypted, encErr := u.cipher.Encrypt(plain)
		if encErr != nil {
			return nil, fmt.Errorf("variableUsecase.Update: encrypt: %w", encErr)
		}
		v.Value = encrypted
	} else {
		v.Value = plain
	}

	now := time.Now().UTC()
	v.ModifiedBy = user
	v.ModifiedAt = &now

	if err := u.repo.Save(ctx, v); err != nil {
		return nil, fmt.Errorf("variableUsecase.Update: %w", err)
	}
	return u.toResponse(v), nil
}

func (u *variableUsecase) plainValue(v *domain.SoarVariable, req dto.UpdateVariableRequest) (string, error) {
	if req.Value != nil && *req.Value != "" && *req.Value != maskedValue {
		return *req.Value, nil
	}
	if !v.IsSecret {
		return v.Value, nil
	}
	plain, err := u.cipher.Decrypt(v.Value)
	if err != nil {
		return "", fmt.Errorf("variableUsecase.Update: decrypt stored value: %w", err)
	}
	return plain, nil
}

func (u *variableUsecase) FindByID(ctx context.Context, id uuid.UUID) (*dto.VariableResponse, error) {
	v, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u.toResponse(v), nil
}

func (u *variableUsecase) FindAll(ctx context.Context, f dto.VariableFilter) ([]dto.VariableResponse, int64, error) {
	items, total, err := u.repo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.VariableResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *u.toResponse(&items[i]))
	}
	return resp, total, nil
}

func (u *variableUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *variableUsecase) InterpolateCommand(ctx context.Context, cmd string) (string, error) {
	matches := variableInterpolationRegex.FindAllStringSubmatch(cmd, -1)
	if len(matches) == 0 {
		return cmd, nil
	}
	names := make([]string, 0, len(matches))
	seen := map[string]bool{}
	for _, m := range matches {
		if !seen[m[1]] {
			names = append(names, m[1])
			seen[m[1]] = true
		}
	}

	vars, err := u.repo.FindByNames(ctx, names)
	if err != nil {
		return cmd, fmt.Errorf("InterpolateCommand: %w", err)
	}

	for _, v := range vars {
		value := v.Value
		if v.IsSecret {
			plain, decErr := u.cipher.Decrypt(value)
			if decErr != nil || plain == "" {
				return cmd, fmt.Errorf("InterpolateCommand: cannot read secret %q", v.Name)
			}
			value = plain
		}
		cmd = strings.ReplaceAll(cmd, "$[variables."+v.Name+"]", value)
	}
	return cmd, nil
}

func (u *variableUsecase) MaskSecrets(ctx context.Context, output string) (string, error) {
	vars, err := u.repo.FindAllPlain(ctx)
	if err != nil {
		return output, fmt.Errorf("MaskSecrets: %w", err)
	}
	for _, v := range vars {
		if !v.IsSecret || v.Value == "" {
			continue
		}
		plain, decErr := u.cipher.Decrypt(v.Value)
		if decErr != nil || plain == "" {
			continue
		}
		output = strings.ReplaceAll(output, plain, strings.Repeat("*", len(plain)))
	}
	return output, nil
}

func (u *variableUsecase) toResponse(v *domain.SoarVariable) *dto.VariableResponse {
	resp := &dto.VariableResponse{
		ID:           v.ID,
		Name:         v.Name,
		Description:  v.Description,
		IsSecret:     v.IsSecret,
		Value:        v.Value,
		CreatedBy:    v.CreatedBy,
		CreatedAt:    v.CreatedAt,
		ModifiedBy:   v.ModifiedBy,
		ModifiedDate: v.ModifiedAt,
	}
	if v.IsSecret {
		resp.Value = maskedValue
	}
	return resp
}
