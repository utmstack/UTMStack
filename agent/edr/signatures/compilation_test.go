package signatures

import (
	"crypto/md5"
	"errors"
	"fmt"
	"testing"
)

func TestContentVersionMatchesCompiledIdentityWithoutParsing(t *testing.T) {
	// Pin the pre-refactor byte framing independently of both public APIs.
	const golden = "go-v1-c736af014292a31c42906a88356232d7c8f16c4dd01c6715d719f7061509971d"
	if version, err := ContentVersion(map[string][]byte{"test.ndb": []byte("Golden.Test:0:*:414243444546\n")}, Limits{}); err != nil || version != golden {
		t.Fatalf("durable identity format changed: version=%q err=%v", version, err)
	}
	files := map[string][]byte{
		"z.ndb": []byte("Remote.Test:0:*:414243444546\n"),
		"a.ndb": []byte("Local.Test:0:*:4748494a4b4c\n"),
	}
	version, err := ContentVersion(files, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	db, err := Compile(files)
	if err != nil || db.Version() != version {
		t.Fatalf("content version %q, compiled version %v, err=%v", version, db, err)
	}
	// Identity checks deliberately do not decide whether source rules are valid.
	// Only the exact known-good content digest can authorize a stored copy.
	files["z.ndb"] = []byte("Remote.Test:0:*:zz\n")
	changed, err := ContentVersion(files, Limits{})
	if err != nil || changed == version {
		t.Fatalf("changed identity=%q, err=%v", changed, err)
	}
	if _, err := Compile(files); err == nil {
		t.Fatal("identity accepted invalid rules as a compiled database")
	}
	for name, tc := range map[string]struct {
		files  map[string][]byte
		limits Limits
	}{
		"path":       {map[string][]byte{"../a.ndb": []byte("x")}, Limits{}},
		"file bytes": {map[string][]byte{"a.ndb": []byte("xx")}, Limits{MaxFileBytes: 1}},
		"total bytes": {map[string][]byte{"a.ndb": []byte("xx"), "b.ndb": []byte("xx")},
			Limits{MaxTotalBytes: 3}},
		"file count": {map[string][]byte{"a.ndb": nil, "b.ndb": nil}, Limits{MaxFiles: 1}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ContentVersion(tc.files, tc.limits); err == nil {
				t.Fatal("identity ignored source validation limits")
			}
		})
	}
}

func TestCompileWithLocalFilesBuildsOnceAndPreservesMergedOrder(t *testing.T) {
	for name, fixture := range map[string]struct {
		base, local map[string][]byte
		data        []byte
	}{
		"body rule order": {
			map[string][]byte{"z.ndb": []byte("Feed.Test:0:*:414243444546\n")},
			map[string][]byte{"a.ndb": []byte("Local.Test:0:*:414243444546\n")},
			[]byte("ABCDEF"),
		},
		"duplicate hash order": {
			map[string][]byte{"z.hdb": []byte(fmt.Sprintf("%x:6:Feed.Test\n", md5.Sum([]byte("ABCDEF"))))},
			map[string][]byte{"a.hdb": []byte(fmt.Sprintf("%x:6:Local.Test\n", md5.Sum([]byte("ABCDEF"))))},
			[]byte("ABCDEF"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			parses, matchers := 0, 0
			db, localErr, fatalErr := compileWithLocalFiles(fixture.base, fixture.local, Limits{}, compileOptions{
				onParse: func() { parses++ }, onFinalize: func() { matchers++ },
			})
			if fatalErr != nil || localErr != nil || db == nil {
				t.Fatalf("local=%v fatal=%v", localErr, fatalErr)
			}
			if parses != 1 || matchers != 1 {
				t.Fatalf("parses=%d matchers=%d; want one of each", parses, matchers)
			}
			assertMatch(t, db, fixture.data, "Local.Test")
		})
	}
}

func TestCompileWithLocalFilesSelectionAndFallback(t *testing.T) {
	base := map[string][]byte{"feed.ndb": []byte("Feed.Test:0:*:414243444546\n")}
	t.Run("global ignore changes feed selection", func(t *testing.T) {
		parses, matchers := 0, 0
		db, localErr, fatalErr := compileWithLocalFiles(base, map[string][]byte{
			"local.ign2": []byte("Feed.Test\n"),
			"local.ndb":  []byte("Local.Test:0:*:4748494a4b4c\n"),
		}, Limits{}, compileOptions{onParse: func() { parses++ }, onFinalize: func() { matchers++ }})
		if localErr != nil || fatalErr != nil || db == nil || parses != 2 || matchers != 2 {
			t.Fatalf("local=%v fatal=%v parses=%d matchers=%d", localErr, fatalErr, parses, matchers)
		}
		assertMatch(t, db, []byte("ABCDEF"), "")
		assertMatch(t, db, []byte("GHIJKL"), "Local.Test")
	})
	t.Run("unchanged local selection keeps single build", func(t *testing.T) {
		parses, matchers := 0, 0
		db, localErr, fatalErr := compileWithLocalFiles(base, map[string][]byte{"local.ign2": []byte("# no selected rules changed\n")},
			Limits{}, compileOptions{onParse: func() { parses++ }, onFinalize: func() { matchers++ }})
		if localErr != nil || fatalErr != nil || db == nil || parses != 1 || matchers != 1 {
			t.Fatalf("local=%v fatal=%v parses=%d matchers=%d", localErr, fatalErr, parses, matchers)
		}
		assertMatch(t, db, []byte("ABCDEF"), "Feed.Test")
	})
	t.Run("global dynamic configuration changes feed selection", func(t *testing.T) {
		parses, matchers := 0, 0
		db, localErr, fatalErr := compileWithLocalFiles(map[string][]byte{
			"feed.ndb": base["feed.ndb"], "feed.pdb": []byte("H:example.com\n"),
		}, map[string][]byte{"local.cfg": []byte("PHISHING:0x0\n")}, Limits{},
			compileOptions{onParse: func() { parses++ }, onFinalize: func() { matchers++ }})
		if localErr != nil || fatalErr != nil || db == nil || parses != 2 || matchers != 2 {
			t.Fatalf("local=%v fatal=%v parses=%d matchers=%d", localErr, fatalErr, parses, matchers)
		}
		if db.DynamicConfig().Phishing != 0 || db.Report().Skipped[SkipDynamicConfig] != 1 {
			t.Fatalf("merged global policy lost: config=%+v skipped=%v", db.DynamicConfig(), db.Report().Skipped)
		}
		assertMatch(t, db, []byte("ABCDEF"), "Feed.Test")
	})
	t.Run("base ignore also selects local rules", func(t *testing.T) {
		db, localErr, fatalErr := CompileWithLocalFiles(map[string][]byte{
			"feed.ndb": base["feed.ndb"], "feed.ign2": []byte("Local.Test\n"),
		}, map[string][]byte{"local.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")}, Limits{})
		if localErr != nil || fatalErr != nil || db == nil {
			t.Fatalf("local=%v fatal=%v", localErr, fatalErr)
		}
		assertMatch(t, db, []byte("ABCDEF"), "Feed.Test")
		assertMatch(t, db, []byte("GHIJKL"), "")
	})
	for name, local := range map[string]map[string][]byte{
		"invalid rule": {"bad.ndb": []byte("Bad.Test:0:*:zz\n"), "good.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")},
		"collision":    {"feed.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")},
		"invalid policy": {"local.cfg": []byte("OTHER:0\n"),
			"local.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")},
	} {
		t.Run(name, func(t *testing.T) {
			db, localErr, fatalErr := CompileWithLocalFiles(base, local, Limits{})
			if db == nil || localErr == nil || fatalErr != nil {
				t.Fatalf("db=%v local=%v fatal=%v", db, localErr, fatalErr)
			}
			assertMatch(t, db, []byte("ABCDEF"), "Feed.Test")
			assertMatch(t, db, []byte("GHIJKL"), "")
		})
	}
}

func TestCompileWithLocalFilesCannotHideInvalidOrDetectionEmptyFeed(t *testing.T) {
	for name, fixture := range map[string]struct{ base, local map[string][]byte }{
		"ignored malformed feed rule": {
			map[string][]byte{"feed.ndb": []byte("Bad.Test:0:*:zz\nGood.Test:0:*:414243444546\n")},
			map[string][]byte{"local.ign2": []byte("Bad.Test\n")},
		},
		"configuration hides malformed feed": {
			map[string][]byte{"feed.pdb": []byte("not a phishing rule\n"), "feed.ndb": []byte("Good.Test:0:*:414243444546\n")},
			map[string][]byte{"local.cfg": []byte("PHISHING:0x0\n")},
		},
		"allowlist only feed": {
			map[string][]byte{"feed.fp": []byte(fmt.Sprintf("%x:6:Allowed.Test\n", md5.Sum([]byte("ABCDEF"))))},
			map[string][]byte{"local.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")},
		},
	} {
		t.Run(name, func(t *testing.T) {
			db, localErr, fatalErr := CompileWithLocalFiles(fixture.base, fixture.local, Limits{})
			if db != nil || localErr != nil || fatalErr == nil {
				t.Fatalf("invalid base accepted: db=%v local=%v fatal=%v", db, localErr, fatalErr)
			}
			if name == "allowlist only feed" && !errors.Is(fatalErr, ErrNoDetections) {
				t.Fatalf("fatal=%v; want ErrNoDetections", fatalErr)
			}
		})
	}
}

func TestCompileWithLocalFilesTracksContainerSourceAndNestedSelection(t *testing.T) {
	feed := map[string][]byte{"daily.cvd": cvdFixture(t, map[string][]byte{
		"feed.ndb": []byte("Feed.Test:0:*:414243444546\n"),
	})}
	for name, local := range map[string]map[string][]byte{
		"plain local": {"local.ndb": []byte("Local.Test:0:*:4748494a4b4c\n")},
		"container policy": {"local.cvd": cvdFixture(t, map[string][]byte{
			"local.ign2": []byte("Feed.Test\n"), "local.ndb": []byte("Local.Test:0:*:4748494a4b4c\n"),
		})},
	} {
		t.Run(name, func(t *testing.T) {
			parses, matchers := 0, 0
			db, localErr, fatalErr := compileWithLocalFiles(feed, local, Limits{},
				compileOptions{onParse: func() { parses++ }, onFinalize: func() { matchers++ }})
			wantParses, wantFeed := 1, "Feed.Test"
			if name == "container policy" {
				wantParses, wantFeed = 2, ""
			}
			if db == nil || localErr != nil || fatalErr != nil || parses != wantParses || matchers != wantParses {
				t.Fatalf("local=%v fatal=%v parses=%d matchers=%d", localErr, fatalErr, parses, matchers)
			}
			assertMatch(t, db, []byte("ABCDEF"), wantFeed)
			assertMatch(t, db, []byte("GHIJKL"), "Local.Test")
		})
	}
}

func TestCompileWithLocalFilesDuplicateProtectedDomainStillCountsAsBaseDetection(t *testing.T) {
	db, localErr, fatalErr := CompileWithLocalFiles(
		map[string][]byte{"z.pdb": []byte("H:example.com\n")},
		map[string][]byte{"a.pdb": []byte("H:example.com\n")}, Limits{})
	if db == nil || localErr != nil || fatalErr != nil || db.DetectionCount() != 1 {
		t.Fatalf("db=%v local=%v fatal=%v", db, localErr, fatalErr)
	}
}
