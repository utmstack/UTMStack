package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type fakeVizRepo struct {
	row   *domain.Visualization
	saved *domain.Visualization
}

func (f *fakeVizRepo) Save(_ context.Context, v *domain.Visualization) error { f.saved = v; return nil }
func (f *fakeVizRepo) FindByID(context.Context, uint64) (*domain.Visualization, error) {
	return f.row, nil
}
func (f *fakeVizRepo) List(context.Context, dto.VisualizationFilter) ([]domain.Visualization, int64, error) {
	return nil, 0, nil
}
func (f *fakeVizRepo) Delete(context.Context, uint64) error { return nil }

var _ connectors.VisualizationRepository = (*fakeVizRepo)(nil)

const goodSpec = `{"dataset":"alerts","chart":"category","dimension":"name","metric":{"agg":"count"}}`

func TestAVisualizationNeedsASpec(t *testing.T) {
	repo := &fakeVizRepo{}
	uc := NewVisualizationUsecase(repo)

	_, err := uc.Create(context.Background(), &domain.Visualization{DashboardID: 1}, "someone")
	if !errors.Is(err, domain.ErrSpecRequired) {
		t.Errorf("err = %v, want ErrSpecRequired", err)
	}
	if repo.saved != nil {
		t.Error("it was stored without a spec")
	}
}

// The spec is validated on the way in, so a widget that could never be answered
// is refused where someone can still fix it.
func TestASpecThatCannotBeAnsweredIsRefusedOnSave(t *testing.T) {
	cases := map[string]string{
		"not json":                   `{`,
		"a dataset that is not ours": `{"dataset":"system.query_log","chart":"metric"}`,
		"a breakdown with nothing to break down by": `{"dataset":"logs","chart":"category"}`,
		"an average with no field":                  `{"dataset":"logs","chart":"metric","metric":{"agg":"avg"}}`,
	}

	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeVizRepo{}
			uc := NewVisualizationUsecase(repo)

			_, err := uc.Create(context.Background(), &domain.Visualization{DashboardID: 1, Spec: spec}, "someone")
			if !errors.Is(err, domain.ErrInvalidSpec) {
				t.Errorf("err = %v, want ErrInvalidSpec", err)
			}
			if repo.saved != nil {
				t.Error("it was stored anyway")
			}
		})
	}
}

func TestAGoodSpecIsStored(t *testing.T) {
	repo := &fakeVizRepo{}
	uc := NewVisualizationUsecase(repo)

	if _, err := uc.Create(context.Background(), &domain.Visualization{DashboardID: 1, Spec: goodSpec}, "someone"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.saved == nil || repo.saved.Spec != goodSpec {
		t.Error("the spec was not stored as given")
	}
}
