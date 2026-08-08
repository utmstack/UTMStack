package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

var variableInterpolationRegex = regexp.MustCompile(`\$\[variables\.([^\]]+)\]`)

type variableUsecase struct {
	repo   connectors.VariableRepository
	cipher connectors.VariableCipher
}

func NewVariableUsecase(repo connectors.VariableRepository, cipher connectors.VariableCipher) connectors.VariableUsecase {
	return &variableUsecase{repo: repo, cipher: cipher}
}

func (u *variableUsecase) Create(ctx context.Context, req dto.CreateVariableRequest, user string) (*dto.VariableResponse, error) {
	now := time.Now().UTC()
	v := &domain.UtmIncidentVariable{
		VariableDescription: req.VariableDescription,
		IsSecret:            req.IsSecret,
		CreatedBy:           &user,
		LastModifiedDate:    &now,
	}
	name := req.VariableName
	v.VariableName = &name

	if req.IsSecret && req.VariableValue != nil && *req.VariableValue != "" {
		encrypted, err := u.cipher.Encrypt(*req.VariableValue)
		if err != nil {
			return nil, fmt.Errorf("variableUsecase.Create: encrypt: %w", err)
		}
		v.VariableValue = &encrypted
	} else {
		v.VariableValue = req.VariableValue
	}

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
	now := time.Now().UTC()
	v.VariableName = req.VariableName
	v.VariableDescription = req.VariableDescription
	v.IsSecret = req.IsSecret
	v.LastModifiedDate = &now
	v.LastModifiedBy = &user

	if req.IsSecret {
		if req.VariableValue != nil && *req.VariableValue != "" && *req.VariableValue != "****" {
			encrypted, encErr := u.cipher.Encrypt(*req.VariableValue)
			if encErr != nil {
				return nil, fmt.Errorf("variableUsecase.Update: encrypt: %w", encErr)
			}
			v.VariableValue = &encrypted
		}
	} else {
		v.VariableValue = req.VariableValue
	}

	if err := u.repo.Save(ctx, v); err != nil {
		return nil, fmt.Errorf("variableUsecase.Update: %w", err)
	}
	return u.toResponse(v), nil
}

func (u *variableUsecase) FindByID(ctx context.Context, id int64) (*dto.VariableResponse, error) {
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

func (u *variableUsecase) Delete(ctx context.Context, id int64) error {
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
	if len(vars) == 0 {
		return cmd, nil
	}

	for _, v := range vars {
		if v.VariableName == nil || v.VariableValue == nil {
			continue
		}
		value := *v.VariableValue
		if v.IsSecret {
			plain, decErr := u.cipher.Decrypt(value)
			if decErr != nil || plain == "" {
				continue
			}
			value = plain
		}
		placeholder := "$[variables." + *v.VariableName + "]"
		cmd = strings.ReplaceAll(cmd, placeholder, value)
	}
	return cmd, nil
}

func (u *variableUsecase) MaskSecrets(ctx context.Context, output string) (string, error) {
	vars, err := u.repo.FindAllPlain(ctx)
	if err != nil {
		return output, fmt.Errorf("MaskSecrets: %w", err)
	}
	for _, v := range vars {
		if !v.IsSecret || v.VariableValue == nil || *v.VariableValue == "" {
			continue
		}
		plain, decErr := u.cipher.Decrypt(*v.VariableValue)
		if decErr != nil || plain == "" {
			continue
		}
		mask := strings.Repeat("*", len(plain))
		output = strings.ReplaceAll(output, plain, mask)
	}
	return output, nil
}

func (u *variableUsecase) toResponse(v *domain.UtmIncidentVariable) *dto.VariableResponse {
	resp := &dto.VariableResponse{
		ID:                  v.ID,
		VariableName:        v.VariableName,
		VariableDescription: v.VariableDescription,
		IsSecret:            v.IsSecret,
		CreatedBy:           v.CreatedBy,
		LastModifiedDate:    v.LastModifiedDate,
		LastModifiedBy:      v.LastModifiedBy,
	}
	if v.IsSecret {
		masked := "****"
		resp.VariableValue = &masked
	} else {
		resp.VariableValue = v.VariableValue
	}
	return resp
}
