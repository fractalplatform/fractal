package relic

import (
	"runtime/debug"
	"testing"
	"text/template"
)

const relicName = "Relic"
const relicURL = "https://github.com/monax/relic"

func TestHistory_DeclareReleases(t *testing.T) {
	_, err := NewHistory(relicName, relicURL).DeclareReleases(
		Release{
			Version: parseVersion(t, "2.1.1"),
			Notes:   `Everything fixed`,
		},
		"2.1.0",
		`Everything broken`,
		"0.0.2",
		`Wonderful things were achieved`,
		"0.0.1",
		`Marvelous advances were made`,
	)
	if err == nil {
		t.Errorf("error expected")
	}

	history, err := NewHistory(relicName, relicURL).DeclareReleases(
		Release{
			Version: parseVersion(t, "2.1.1"),
			Notes:   `Everything fixed`,
		},
		"2.1.0",
		`Everything broken`,
		"2.0.0",
		`Wonderful things were achieved`,
		"1.0.0",
		`Wonderful things were achieved`,
		"0.0.2",
		`Wonderful things were achieved`,
		"0.0.1",
		`Marvelous advances were made`,
	)
	if err != nil {
		t.Error(err)
	}
	changelog, err := history.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [2.1.1]\nEverything fixed\n\n## [2.1.0]\nEverything broken\n\n## [2.0.0]\nWonderful things were achieved\n\n## [1.0.0]\nWonderful things were achieved\n\n## [0.0.2]\nWonderful things were achieved\n\n## [0.0.1]\nMarvelous advances were made\n\n[2.1.1]: https://github.com/monax/relic/compare/v2.1.0...v2.1.1\n[2.1.0]: https://github.com/monax/relic/compare/v2.0.0...v2.1.0\n[2.0.0]: https://github.com/monax/relic/compare/v1.0.0...v2.0.0\n[1.0.0]: https://github.com/monax/relic/compare/v0.0.2...v1.0.0\n[0.0.2]: https://github.com/monax/relic/compare/v0.0.1...v0.0.2\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)

	// Fail gap
	_, err = NewHistory(relicName, relicURL).DeclareReleases(
		Release{
			Version: parseVersion(t, "1.0.3"),
			Notes:   `Wonderful things were achieved`,
		},
		"0.0.2",
		`Wonderful things were achieved`,
		Release{
			Version: parseVersion(t, "0.0.1"),
			Notes:   `Marvelous advances were made`,
		},
	)
	if err == nil {
		t.Errorf("error expected")
	}

	history, err = NewHistory(relicName, relicURL).DeclareReleases(
		Release{
			Version: parseVersion(t, "1.0.3"),
			Notes:   `Wonderful things were achieved`,
		},
		"1.0.2",
		`Hotfix`,
		"1.0.1",
		`Hotfix`,
		"1.0.0",
		`Wonderful things were achieved`,
		Release{
			Version: parseVersion(t, "0.0.1"),
			Notes:   `Marvelous advances were made`,
		},
	)
	if err != nil {
		t.Error(err)
	}
	changelog, err = history.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [1.0.3]\nWonderful things were achieved\n\n## [1.0.2]\nHotfix\n\n## [1.0.1]\nHotfix\n\n## [1.0.0]\nWonderful things were achieved\n\n## [0.0.1]\nMarvelous advances were made\n\n[1.0.3]: https://github.com/monax/relic/compare/v1.0.2...v1.0.3\n[1.0.2]: https://github.com/monax/relic/compare/v1.0.1...v1.0.2\n[1.0.1]: https://github.com/monax/relic/compare/v1.0.0...v1.0.1\n[1.0.0]: https://github.com/monax/relic/compare/v0.0.1...v1.0.0\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)

	_, err = NewHistory(relicName, relicURL).DeclareReleases(
		"0.1.3",
		`Wonderful things were achieved`,
		"0.0.2",
		`Wonderful things were achieved`,
		"0.0.1",
		`Marvelous advances were made`,
	)
	if err == nil {
		t.Errorf("error expected")
	}

	history, err = NewHistory(relicName, relicURL).DeclareReleases(
		"0.0.3",
		`Wonderful things were achieved`,
		"0.0.2",
		`Wonderful things were achieved`,
		"0.0.1",
		`Marvelous advances were made`,
	)
	if err != nil {
		t.Error(err)
	}
	changelog, err = history.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [0.0.3]\nWonderful things were achieved\n\n## [0.0.2]\nWonderful things were achieved\n\n## [0.0.1]\nMarvelous advances were made\n\n[0.0.3]: https://github.com/monax/relic/compare/v0.0.2...v0.0.3\n[0.0.2]: https://github.com/monax/relic/compare/v0.0.1...v0.0.2\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)

	_, err = NewHistory(relicName, relicURL).DeclareReleases(
		"0.0.3",
		`Wonderful things were achieved`,
		"0.0.2",
		`Wonderful things were achieved`,
		"0.0.1",
	)
	if err == nil {
		t.Errorf("error expected")
	}

	_, err = NewHistory(relicName, relicURL).DeclareReleases(
		"0.0.2",
		`Wonderful things were achieved`,
		"0.0.3",
		`Wonderful things were achieved`,
		"0.0.1",
		`Marvelous advances were made`,
	)
	if err == nil {
		t.Errorf("error expected")
	}

	_, err = history.DeclareReleases(
		"1.0.0",
		"finally",
		"0.2.1",
		"",
		"0.2.0",
		"Came after 0.1.0",
	)
	if err == nil {
		t.Errorf("error expected")
	}
}

func TestHistory_DeclareReleases_Multiple(t *testing.T) {
	history, err := NewHistory(relicName, relicURL).DeclareReleases(
		"0.1.0",
		"Basic functionality",
		"0.0.2",
		"Build scripts",
		"0.0.1",
		"Proof of concept",
	)
	if err != nil {
		t.Error(err)
	}
	changelog, err := history.Changelog()
	if err != nil {
		t.Error(err)
	}

	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [0.1.0]\nBasic functionality\n\n## [0.0.2]\nBuild scripts\n\n## [0.0.1]\nProof of concept\n\n[0.1.0]: https://github.com/monax/relic/compare/v0.0.2...v0.1.0\n[0.0.2]: https://github.com/monax/relic/compare/v0.0.1...v0.0.2\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)

	history1, err := history.DeclareReleases(
		"1.0.0",
		"finally",
		"0.2.1",
		"Patch",
		"0.2.0",
		"Came after 0.1.0",
	)
	if err != nil {
		t.Error(err)
	}
	if history1 != history {
		t.Errorf("history1 and history should be a pointer to the same object")
	}

	changelog, err = history1.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [1.0.0]\nfinally\n\n## [0.2.1]\nPatch\n\n## [0.2.0]\nCame after 0.1.0\n\n## [0.1.0]\nBasic functionality\n\n## [0.0.2]\nBuild scripts\n\n## [0.0.1]\nProof of concept\n\n[1.0.0]: https://github.com/monax/relic/compare/v0.2.1...v1.0.0\n[0.2.1]: https://github.com/monax/relic/compare/v0.2.0...v0.2.1\n[0.2.0]: https://github.com/monax/relic/compare/v0.1.0...v0.2.0\n[0.1.0]: https://github.com/monax/relic/compare/v0.0.2...v0.1.0\n[0.0.2]: https://github.com/monax/relic/compare/v0.0.1...v0.0.2\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)
	_, err = history.DeclareReleases(
		"0.1.3",
		`New newness`,
		"0.1.2",
		`Added blockchain`,
		"0.1.1",
		"Dried apricot",
	)
	if err == nil {
		t.Errorf("error expected")
	}
}

func TestHistory_WithChangelogTemplate(t *testing.T) {
	history, err := NewHistory("Test Project", relicURL).
		WithChangelogTemplate(template.Must(template.New("tests").
			Parse("{{range .Releases}}{{$.Project}} (v{{.Version}}): {{.Notes}}\n{{end}}"))).
		DeclareReleases(
			"0.1.0",
			"Basic functionality",
			"0.0.2",
			"Build scripts",
			"0.0.1",
			"Proof of concept",
		)
	if err != nil {
		t.Fatal(err)
	}
	changelog, err := history.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "Test Project (v0.1.0): Basic functionality\nTest Project (v0.0.2): Build scripts\nTest Project (v0.0.1): Proof of concept\n",
		changelog)
}

func TestHistory_Release(t *testing.T) {
	history, err := NewHistory(relicName, relicURL).
		DeclareReleases(
			"0.1.0",
			"Basic functionality",
			"0.0.2",
			"Build scripts",
			"0.0.1",
			"Proof of concept",
		)
	if err != nil {
		t.Fatal(err)
	}

	release, err := history.Release("0.0.2")
	if err != nil {
		t.Fatal(err)
	}
	if release.Notes != "Build scripts" {
		t.Errorf("release notes should be 'Build scripts' but is '%s'", release.Notes)
	}
}

func TestHistory_Changelog_Dates(t *testing.T) {
	history, err := NewHistory(relicName, relicURL).
		DeclareReleases("",
			"Some unreleased things",
			"0.1.0 - 2018-08-01",
			"Basic functionality",
			"0.0.2",
			"Build scripts",
			"0.0.1 - 2018-02-28",
			"Proof of concept",
		)
	if err != nil {
		t.Fatal(err)
	}
	changelog, err := history.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [Unreleased]\nSome unreleased things\n\n## [0.1.0] - 2018-08-01\nBasic functionality\n\n## [0.0.2]\nBuild scripts\n\n## [0.0.1] - 2018-02-28\nProof of concept\n\n[Unreleased]: https://github.com/monax/relic/compare/v0.1.0...HEAD\n[0.1.0]: https://github.com/monax/relic/compare/v0.0.2...v0.1.0\n[0.0.2]: https://github.com/monax/relic/compare/v0.0.1...v0.0.2\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)
}

func TestHistory_Changelog_Unreleased(t *testing.T) {
	history, err := NewHistory(relicName, relicURL).
		DeclareReleases("",
			"Some unreleased things",
		)
	if err != nil {
		t.Fatal(err)
	}
	changelog, err := history.Changelog()
	if err != nil {
		t.Error(err)
	}

	_, err = NewHistory(relicName, relicURL).
		DeclareReleases("",
			"Some unreleased things",
			"",
			"more unreleased things",
		)
	if err == nil {
		t.Fatal("Should not allow multiple unreleased sections")
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [Unreleased]\nSome unreleased things\n\n[Unreleased]: https://github.com/monax/relic/commits/HEAD",
		changelog)

	_, err = NewHistory(relicName, relicURL).
		DeclareReleases("",
			"Some unreleased things",
			"0.0.1",
			"Initial version",
			"",
			"more unreleased things",
		)
	if err == nil {
		t.Fatal("Should not allow multiple unreleased sections")
	}

	history, err = NewHistory(relicName, relicURL).
		DeclareReleases("",
			"Some unreleased things",
			"1.0.0",
			"More to add",
			"0.0.1",
			"Initial version",
		)
	changelog, err = history.Changelog()
	if err != nil {
		t.Error(err)
	}
	assertChangelog(t, "# [Relic](https://github.com/monax/relic) Changelog\n## [Unreleased]\nSome unreleased things\n\n## [1.0.0]\nMore to add\n\n## [0.0.1]\nInitial version\n\n[Unreleased]: https://github.com/monax/relic/compare/v1.0.0...HEAD\n[1.0.0]: https://github.com/monax/relic/compare/v0.0.1...v1.0.0\n[0.0.1]: https://github.com/monax/relic/commits/v0.0.1",
		changelog)
}

func assertChangelog(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("expected changelog:\n%s\n\nBut actual changelog was:\n\n%s\nActual (yankable):\nassertChangelog(t, %#v,\nchangelog)\n\n%s\n",
			expected, actual, actual, debug.Stack())
	}
}

func parseVersion(t *testing.T, versionString string) Version {
	version, err := ParseVersion(versionString)
	if err != nil {
		t.Error(err)
	}
	return version
}
