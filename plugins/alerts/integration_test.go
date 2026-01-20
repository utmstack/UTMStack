package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAlertsIntegration(t *testing.T) {
	// 1. Setup OpenSearch connection
	nodesEnv := os.Getenv("NODES")
	if nodesEnv == "" {
		t.Skip("NODES env var not set, skipping integration test")
	}

	// Fix for http vs https mismatch in test environment
	if strings.HasPrefix(nodesEnv, "https://") {
		nodesEnv = strings.Replace(nodesEnv, "https://", "http://", 1)
	}
	if !strings.HasPrefix(nodesEnv, "http://") && !strings.HasPrefix(nodesEnv, "https://") {
		nodesEnv = "http://" + nodesEnv
	}

	err := sdkos.Connect([]string{nodesEnv}, os.Getenv("USER"), os.Getenv("PASSWORD"))
	if err != nil {
		t.Fatalf("Failed to connect to OpenSearch: %v", err)
	}

	// Helper to create a test alert
	createAlert := func(name, user string, dedup []string) *plugins.Alert {
		return &plugins.Alert{
			Id:          uuid.NewString(),
			Name:        name,
			Description: "Integration Test Alert",
			Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
			Severity:    "low",
			Events: []*plugins.Event{
				{
					Log: map[string]*structpb.Value{
						"user": structpb.NewStringValue(user),
					},
				},
			},
			DeduplicateBy: dedup,
			GroupBy:       dedup, // Default to using same fields for GroupBy in legacy tests
		}
	}

	// Helper to clean up
	alertIndex := sdkos.BuildIndexPattern("v11", "alert")
	runID := uuid.NewString()

	t.Run("Deduplication_Basic", func(t *testing.T) {
		alertName := "DedupTest-" + runID
		dedupFields := []string{"name"}

		// 1. Send First Alert
		alert1 := createAlert(alertName, "user1", dedupFields)
		_, err := correlate(context.Background(), alert1)
		if err != nil {
			t.Fatalf("First correlate failed: %v", err)
		}

		// Allow OS to index
		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// 2. Send Second Alert (DIFFERENT user so it's not a duplicate, but SAME name so it's grouped)
		alert2 := createAlert(alertName, "user2", dedupFields)
		alert2.DeduplicateBy = []string{"name", "events.0.log.user"}
		alert2.GroupBy = []string{"name"}

		_, err = correlate(context.Background(), alert2)
		if err != nil {
			t.Fatalf("Second correlate failed: %v", err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// Fetch alert2
		a2Doc, err := getAlertFromOS(alert2.Id)
		if err != nil {
			t.Fatalf("Failed to fetch alert 2: %v", err)
		}

		if a2Doc.ParentID == nil {
			t.Errorf("Expected Alert 2 to have ParentID, got nil")
		} else if *a2Doc.ParentID != alert1.Id {
			t.Errorf("Expected Alert 2 ParentID to be %s, got %s", alert1.Id, *a2Doc.ParentID)
		}
	})

	t.Run("Deduplication_GroupBy_Fields", func(t *testing.T) {
		alertName := "GroupByTest-" + runID
		groupBy := []string{"name"}
		dedupFields := []string{"name", "events.0.log.user"}

		// 1. User A - First Alert
		alertA1 := createAlert(alertName, "userA", dedupFields)
		alertA1.GroupBy = groupBy

		_, err := correlate(context.Background(), alertA1)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// 2. User A - Second Alert -> DUPLICATE -> Should be dropped
		alertA2_dup := createAlert(alertName, "userA", dedupFields)
		alertA2_dup.GroupBy = groupBy
		_, err = correlate(context.Background(), alertA2_dup)
		if err != nil {
			t.Fatal(err)
		}

		// 3. User B - First Alert -> DIFFERENT for Dedup, SAME for GroupBy -> Should be linked
		alertB1 := createAlert(alertName, "userB", dedupFields)
		alertB1.GroupBy = groupBy
		_, err = correlate(context.Background(), alertB1)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// Verify A2_dup is NOT in OS
		_, err = getAlertFromOS(alertA2_dup.Id)
		if err == nil {
			t.Errorf("Expected A2_dup to be dropped")
		}

		// Verify B1 has parent A1
		docB1, err := getAlertFromOS(alertB1.Id)
		if err != nil {
			t.Fatalf("Failed to get B1: %v", err)
		}
		if docB1.ParentID == nil || *docB1.ParentID != alertA1.Id {
			t.Errorf("Expected B1 parent to be A1. Got: %v", docB1.ParentID)
		}
	})

	// Helper to get hit from OS
	getHit := func(id string) (*sdkos.Hit, error) {
		query := sdkos.SearchRequest{
			Query: &sdkos.Query{
				Term: map[string]map[string]interface{}{
					"id.keyword": {
						"value": id,
					},
				},
			},
			Size: 1,
		}
		hits, err := query.WideSearchIn(context.Background(), []string{sdkos.BuildIndexPattern("v11", "alert")})
		if err != nil {
			return nil, err
		}
		if hits.Hits.Total.Value == 0 {
			return nil, fmt.Errorf("not found")
		}
		return &hits.Hits.Hits[0], nil
	}

	t.Run("Deduplication_Reopen_Closed", func(t *testing.T) {
		alertName := "ReopenTest-" + runID
		dedupFields := []string{"name"}

		// 1. Create Parent Alert
		alertParent := createAlert(alertName, "user1", dedupFields)
		_, err := correlate(context.Background(), alertParent)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// 2. Manually Close the Parent Alert
		hit, err := getHit(alertParent.Id)
		if err != nil {
			t.Fatal(err)
		}

		var a AlertFields
		hit.Source.ParseSource(&a)
		a.Status = 5
		hit.Source.SetSource(a)

		err = hit.Save(context.Background())
		if err != nil {
			t.Fatal("Failed to save closed alert:", err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// 3. Create Child Alert (DIFFERENT user so it's not a duplicate, but SAME name so it's grouped)
		alertChild := createAlert(alertName, "user2", dedupFields)
		alertChild.DeduplicateBy = []string{"name", "events.0.log.user"}
		alertChild.GroupBy = []string{"name"}
		_, err = correlate(context.Background(), alertChild)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		// 4. Verify Parent is Re-opened (Status 2)
		docParentReopened, err := getAlertFromOS(alertParent.Id)
		if err != nil {
			t.Fatal(err)
		}

		if docParentReopened.Status != 2 {
			t.Errorf("Expected Parent Alert to be re-opened (Status 2), but got %d", docParentReopened.Status)
		}

		// Verify Grouping
		docChild, _ := getAlertFromOS(alertChild.Id)
		if docChild.ParentID == nil || *docChild.ParentID != alertParent.Id {
			t.Errorf("Expected Child to group with Parent %s, got %v", alertParent.Id, docChild.ParentID)
		}
	})

	t.Run("Deduplication_Missing_Field", func(t *testing.T) {
		alertName := "MissingFieldTest-" + runID
		dedupFields := []string{"name", "events.0.log.user"}

		alertNoUser := &plugins.Alert{
			Id:            uuid.NewString(),
			Name:          alertName,
			Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			Severity:      "low",
			Events:        []*plugins.Event{{Log: map[string]*structpb.Value{}}},
			DeduplicateBy: dedupFields,
			GroupBy:       dedupFields,
		}

		alertWithUser := createAlert(alertName, "bob", dedupFields)

		_, err := correlate(context.Background(), alertNoUser)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		_, err = correlate(context.Background(), alertWithUser)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		docWithUser, _ := getAlertFromOS(alertWithUser.Id)
		if docWithUser.ParentID != nil && *docWithUser.ParentID == alertNoUser.Id {
			t.Errorf("Alert with User='bob' incorrectly deduped with Alert (User=Missing)")
		}

		alertNoUser2 := &plugins.Alert{
			Id:            uuid.NewString(),
			Name:          alertName,
			Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			Severity:      "low",
			Events:        []*plugins.Event{{Log: map[string]*structpb.Value{"diff": structpb.NewStringValue("yes")}}},
			DeduplicateBy: dedupFields,
			GroupBy:       []string{"name"},
		}

		_, err = correlate(context.Background(), alertNoUser2)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		docNoUser2, err := getAlertFromOS(alertNoUser2.Id)
		if err != nil {
			t.Logf("Alert(NoUser2) correctly dropped or not found yet: %v", err)
		} else if docNoUser2.ParentID == nil {
			t.Errorf("Expected Alert(NoUser2) to find a parent")
		}
	})

	t.Run("Deduplication_Empty_List", func(t *testing.T) {
		alertName := "EmptyDedup-" + runID
		alert1 := createAlert(alertName, "user1", []string{})
		_, err := correlate(context.Background(), alert1)
		if err != nil {
			t.Fatal(err)
		}
		alert2 := createAlert(alertName, "user1", []string{})
		_, err = correlate(context.Background(), alert2)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)
		doc1, _ := getAlertFromOS(alert1.Id)
		doc2, _ := getAlertFromOS(alert2.Id)
		if doc1 != nil && doc1.ParentID != nil {
			t.Errorf("Alert 1 should be parent")
		}
		if doc2 != nil && doc2.ParentID != nil {
			t.Errorf("Alert 2 should be separate parent")
		}
	})

	t.Run("Deduplication_vs_GroupBy", func(t *testing.T) {
		alertName := "DedupVsGroup-" + runID
		groupBy := []string{"name"}
		dedupBy := []string{"name", "events.0.log.user"}

		alert1 := &plugins.Alert{
			Id:            uuid.NewString(),
			Name:          alertName,
			Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			Events:        []*plugins.Event{{Log: map[string]*structpb.Value{"user": structpb.NewStringValue("userA")}}},
			DeduplicateBy: dedupBy,
			GroupBy:       groupBy,
		}
		_, err := correlate(context.Background(), alert1)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		alert2 := &plugins.Alert{
			Id:            uuid.NewString(),
			Name:          alertName,
			Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			Events:        []*plugins.Event{{Log: map[string]*structpb.Value{"user": structpb.NewStringValue("userA")}}},
			DeduplicateBy: dedupBy,
			GroupBy:       groupBy,
		}
		_, err = correlate(context.Background(), alert2)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)

		_, err = getAlertFromOS(alert2.Id)
		if err == nil {
			t.Errorf("Alert 2 (Duplicate) should not have been indexed")
		}

		alert3 := &plugins.Alert{
			Id:            uuid.NewString(),
			Name:          alertName,
			Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
			Events:        []*plugins.Event{{Log: map[string]*structpb.Value{"user": structpb.NewStringValue("userB")}}},
			DeduplicateBy: dedupBy,
			GroupBy:       groupBy,
		}
		_, err = correlate(context.Background(), alert3)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
		sdkos.RefreshIndex(context.Background(), alertIndex)
		doc3, err := getAlertFromOS(alert3.Id)
		if err != nil {
			t.Fatalf("Failed to get Alert 3: %v", err)
		}
		if doc3.ParentID == nil || *doc3.ParentID != alert1.Id {
			t.Errorf("Alert 3 should have been grouped with Alert 1. Got: %v", doc3.ParentID)
		}
	})
}

func getAlertFromOS(id string) (*AlertFields, error) {
	query := sdkos.SearchRequest{
		Query: &sdkos.Query{
			Term: map[string]map[string]interface{}{
				"id.keyword": {
					"value": id,
				},
			},
		},
		Size: 1,
	}
	ctx := context.Background()
	hits, err := query.WideSearchIn(ctx, []string{sdkos.BuildIndexPattern("v11", "alert")})
	if err != nil {
		return nil, err
	}
	if hits.Hits.Total.Value == 0 {
		return nil, fmt.Errorf("alert not found")
	}
	var a AlertFields
	err = hits.Hits.Hits[0].Source.ParseSource(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
