package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"

	"github.com/google/uuid"
)

type fakeVizRepo struct {
	row   *domain.Visualization
	saved *domain.Visualization
}

// someID stands for any existing row; the tests care about ownership, not
// which identifier it has.
var someID = uuid.MustParse("8f1c1b8e-0000-4000-8000-00000000000a")

func (f *fakeVizRepo) Save(_ context.Context, v *domain.Visualization) error { f.saved = v; return nil }
func (f *fakeVizRepo) FindByID(context.Context, uuid.UUID) (*domain.Visualization, error) {
	return f.row, nil
}
func (f *fakeVizRepo) List(context.Context, dto.VisualizationFilter) ([]domain.Visualization, int64, error) {
	return nil, 0, nil
}
func (f *fakeVizRepo) Delete(context.Context, uuid.UUID) error { return nil }

var _ connectors.VisualizationRepository = (*fakeVizRepo)(nil)

const goodSpec = `{"dataset":"alerts","chart":"category","dimension":"name","metric":{"agg":"count"}}`

func TestAVisualizationNeedsASpec(t *testing.T) {
	repo := &fakeVizRepo{}
	uc := NewVisualizationUsecase(repo)

	_, err := uc.Create(context.Background(), &domain.Visualization{DashboardID: someID}, "someone")
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

			_, err := uc.Create(context.Background(), &domain.Visualization{DashboardID: someID, Spec: spec}, "someone")
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

	if _, err := uc.Create(context.Background(), &domain.Visualization{DashboardID: someID, Spec: goodSpec}, "someone"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.saved == nil || repo.saved.Spec != goodSpec {
		t.Error("the spec was not stored as given")
	}
}

// The four shapes the editor can build. It saves the spec as JSON text, so this
// is exactly what arrives on the wire — if one of them stops validating, that
// chart type stops being saveable.
func TestEveryWidgetTheEditorBuildsIsAccepted(t *testing.T) {
	specs := map[string]string{
		"a single number": `{"dataset":"logs","chart":"metric","metric":{"agg":"count"}}`,
		"top values of a field": `{"dataset":"alerts","chart":"category","metric":{"agg":"count"},` +
			`"dimension":"name","limit":5}`,
		"a timeline split by a field": `{"dataset":"logs","chart":"time","metric":{"agg":"count"},` +
			`"dimension":"dataType","interval":"1d"}`,
		"records in a table": `{"dataset":"alerts","chart":"table","metric":{"agg":"count"},` +
			`"columns":["name","severity"],"filters":[{"field":"severity","op":"not_eq","value":"low"}]}`,
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			repo := &fakeVizRepo{}
			uc := NewVisualizationUsecase(repo)
			_, err := uc.Create(context.Background(), &domain.Visualization{
				DashboardID: someID, Spec: spec,
			}, "someone")
			if err != nil {
				t.Fatalf("refused: %v", err)
			}
		})
	}
}

// The store counts records and nothing else, so a widget asking for an average
// is refused at the door rather than saved and answered with a count.
func TestAMeasureTheStoreCannotAnswerIsRefused(t *testing.T) {
	repo := &fakeVizRepo{}
	uc := NewVisualizationUsecase(repo)

	_, err := uc.Create(context.Background(), &domain.Visualization{
		DashboardID: someID,
		Spec:        `{"dataset":"logs","chart":"metric","metric":{"agg":"avg","field":"bytes"}}`,
	}, "someone")
	if !errors.Is(err, domain.ErrInvalidSpec) {
		t.Fatalf("err = %v, want an invalid spec", err)
	}
	if repo.saved != nil {
		t.Error("it was saved anyway")
	}
}
