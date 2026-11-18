package signatures

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestPasswordDatabaseSelectionAndOrdering(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"z.ign2": []byte("Skip.Password\n"),
		"a.pwdb": []byte("Any.Password;Engine:81-255,Container:CL_TYPE_ANY;0;any\n" +
			"Skip.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;skip\n" +
			"Old.Password;Engine:81-100,Container:CL_TYPE_ZIP;0;old\n" +
			"Zip.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;zip\n" +
				"Later.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;later\n" +
				"NoEngine.Password;Target:0;0;noengine\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	d.ForEachZIPPassword(func(p []byte) bool { got = append(got, string(p)); return false })
	if len(got) != 4 || got[0] != "later" || got[1] != "zip" || got[2] != "noengine" || got[3] != "any" {
		t.Fatalf("password selection/order = %q", got)
	}
	if d.Report().Skipped[SkipIgnoreList] != 1 || d.Report().Skipped[SkipLevel] != 1 {
		t.Fatalf("password skips: %+v", d.Report().Skipped)
	}
}

func TestPasswordDatabaseRejectsMalformedEntries(t *testing.T) {
	for _, line := range []string{
		"Missing.Fields", "Bad;Engine:81-255,Container:CL_TYPE_ZIP;2;secret",
		"Bad;Engine:81-255,Container:CL_TYPE_ZIP;1;xyz",
		"Bad;;0;secret",
		"Bad;Engine:81-255,FileSize:12;0;secret",
	} {
		if _, err := Compile(map[string][]byte{"x.pwdb": []byte(line + "\n")}); err == nil {
			t.Errorf("accepted malformed password entry %q", line)
		}
	}
}

func TestPasswordDatabaseInsideCVDAndNameOnlyIgnoreDigest(t *testing.T) {
	name := "Skip.Password"
	archive := cvdFixtureOrdered(t, []member{
		{"daily.ign2", []byte(fmt.Sprintf("%s:%x\n", name, md5.Sum([]byte(name))))},
		{"daily.pwdb", []byte("Skip.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;ignored\n" +
			"Storage.Future;Engine:81-255,Container:CL_TYPE_ZIP;2;ignored\n" +
			"Keep.Password;Engine:81-255,Container:CL_TYPE_ZIP;1;736563726574\n")},
		{"daily.ndb", []byte("Marker:0:*:41424344\n")},
	})
	d, err := Compile(map[string][]byte{"daily.cvd": archive})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	d.ForEachZIPPassword(func(p []byte) bool { got = append(got, string(p)); return false })
	if len(got) != 1 || got[0] != "secret" {
		t.Fatalf("passwords from CVD = %q", got)
	}
	if d.Report().Skipped[SkipIgnoreList] != 1 || d.Report().Skipped[SkipPasswordStorage] != 1 {
		t.Fatalf("password skips: %+v", d.Report().Skipped)
	}
}
