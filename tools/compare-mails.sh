#!/usr/bin/env bash
#
# Compare the mails (HTML and text part) rendered by the current git index state
# with those rendered by a given git revision.
#
# Both versions are exported to a work directory (the repository itself is not touched),
# get the same downloaded copy of each feed as input, and render every item with the
# respective templates. Afterwards, the outputs are compared file by file.
#
# Usage: compare-mails.sh [-f feeds-file] [-o work-dir] [-i] [-q] <revision>
#
#   -f feeds-file  File with one feed per line: '<name> <url>' (# starts a comment).
#                  Defaults to the built-in list below.
#   -o work-dir    Directory to use for downloads, exported trees, and output.
#                  Defaults to a fresh temporary directory, which is kept afterwards.
#   -i             Include images (downloads them). Without, the body HTML is not
#                  re-serialized, so changes in that processing go unnoticed.
#   -q             Only print the summary, not the diffs of differing files.
#
# Exit status: 0 if all mails are equal (ignoring whitespace and blank lines, as 'diff -w -B'),
#              1 if some differ, 2 on errors.
# Needs network access for downloading the feeds (and images with -i).

set -euo pipefail

DEFAULT_FEEDS='
xkcd      https://xkcd.com/atom.xml
spiegel   https://www.spiegel.de/schlagzeilen/index.rss
smbc      https://www.smbc-comics.com/comic/rss
gonews    https://go.dev/blog/feed.atom
ghcommits https://github.com/Necoro/feed2imap-go/commits/master.atom
lemmy     https://lemmy.world/feeds/c/technology.xml?sort=New
mastodon  https://mastodon.social/@Mastodon.rss
jsonfeed  https://www.jsonfeed.org/feed.json
heise     https://www.heise.de/rss/heise-atom.xml
'

usage() {
    sed -n '/^# Usage:/,/^$/s/^# \{0,1\}//p' "$0" >&2
    exit 2
}

feeds_file=""
work_dir=""
images=""
quiet=0
while getopts "f:o:iqh" opt; do
    case $opt in
        f) feeds_file=$OPTARG ;;
        o) work_dir=$OPTARG ;;
        i) images=1 ;;
        q) quiet=1 ;;
        *) usage ;;
    esac
done
shift $((OPTIND - 1))
[[ $# -eq 1 ]] || usage
rev=$1

repo=$(git rev-parse --show-toplevel)
git -C "$repo" rev-parse --verify --quiet "$rev^{commit}" >/dev/null || { echo "Unknown revision: $rev" >&2; exit 2; }

if [[ -n $feeds_file ]]; then
    feeds=$(<"$feeds_file")
else
    feeds=$DEFAULT_FEEDS
fi

if [[ -z $work_dir ]]; then
    work_dir=$(mktemp -d -t f2i-compare-mails.XXXXXX)
else
    mkdir -p "$work_dir"
    work_dir=$(realpath "$work_dir")
fi
rm -rf "$work_dir"/{feeds,old,new,out-old,out-new}
mkdir -p "$work_dir"/{feeds,old,new}

echo ">>> Work directory: $work_dir"

# --- Download feeds (once, so both versions get identical input) ---
names=()
while read -r name url _; do
    [[ -z $name || $name == \#* ]] && continue
    echo ">>> Fetching $name from $url"
    if curl -fsSL -A "Mozilla/5.0 (feed2imap-go compare-mails)" -o "$work_dir/feeds/$name.xml" "$url"; then
        names+=("$name")
    else
        echo "!!! Failed to fetch $name, skipping." >&2
    fi
done <<< "$feeds"
[[ ${#names[@]} -gt 0 ]] || { echo "No feeds fetched." >&2; exit 2; }

# --- Export trees ---
echo ">>> Exporting $rev"
git -C "$repo" archive "$rev" | tar -x -C "$work_dir/old"
echo ">>> Exporting index"
git -C "$repo" checkout-index -a --prefix="$work_dir/new/"

# --- Render ---
cat > "$work_dir/zz_compare_mails_test.go" <<'EOF'
package feed

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Necoro/feed2imap-go/internal/feed/template"
	"github.com/Necoro/feed2imap-go/pkg/config"
)

func TestCompareMailsRender(t *testing.T) {
	in, out := os.Getenv("F2I_FEEDS"), os.Getenv("F2I_OUT")

	files, err := filepath.Glob(filepath.Join(in, "*.xml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".xml")
		opts := config.DefaultFeedOptions
		opts.InclImages = os.Getenv("F2I_IMAGES") != ""
		cfg := &config.Feed{Name: name, Exec: []string{"cat", file}, Options: opts}

		f, err := Create(cfg, config.DefaultGlobalOptions)
		if err != nil {
			t.Fatal(err)
		}
		if err = f.Parse(); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		dir := filepath.Join(out, name)
		if err = os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}

		for i := range f.items {
			item := &f.items[i]
			item.buildBody()
			for ext, tpl := range map[string]template.Template{"html": template.Html, "txt": template.Text} {
				var b bytes.Buffer
				if err = tpl.Execute(&b, item); err != nil {
					t.Errorf("%s/%d (%s): %v", name, i, ext, err)
					continue
				}
				if err = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%02d.%s", i, ext)), b.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
EOF

for v in old new; do
    echo ">>> Rendering $v"
    cp "$work_dir/zz_compare_mails_test.go" "$work_dir/$v/internal/feed/"
    F2I_FEEDS="$work_dir/feeds" F2I_OUT="$work_dir/out-$v" F2I_IMAGES=$images \
        go test -C "$work_dir/$v" -vet=off -count=1 -run '^TestCompareMailsRender$' ./internal/feed/ > "$work_dir/render-$v.log" 2>&1 ||
        { echo "!!! Rendering $v failed:" >&2; cat "$work_dir/render-$v.log" >&2; exit 2; }
done

# --- Compare ---
echo
echo ">>> Comparing $rev (old) with index (new)"
printf '%-12s %-5s %6s %6s %6s %6s\n' feed part total same ws-only differ
diffs=()
for name in "${names[@]}"; do
    for ext in html txt; do
        same=0 ws=0 differ=0
        # union of files in both outputs, so that added/removed items are detected
        for file in $(ls "$work_dir/out-old/$name" "$work_dir/out-new/$name" 2>/dev/null | grep "\.$ext\$" | sort -u); do
            o="$work_dir/out-old/$name/$file" n="$work_dir/out-new/$name/$file"
            if cmp -s "$o" "$n"; then
                same=$((same + 1))
            elif [[ -e $o && -e $n ]] && diff -q -w -B "$o" "$n" >/dev/null; then
                ws=$((ws + 1))
            else
                differ=$((differ + 1)); diffs+=("$name/$file")
            fi
        done
        printf '%-12s %-5s %6d %6d %6d %6d\n' "$name" "$ext" $((same + ws + differ)) $same $ws $differ
    done
done

if [[ $quiet -eq 0 ]]; then
    for file in "${diffs[@]}"; do
        echo
        diff -u -w -B --label "old/$file" --label "new/$file" \
            "$work_dir/out-old/$file" "$work_dir/out-new/$file" 2>&1 || true
    done
fi

echo
echo ">>> Outputs: $work_dir/out-old  $work_dir/out-new"
echo ">>> Full diff: diff -ru '$work_dir/out-old' '$work_dir/out-new'"
[[ ${#diffs[@]} -eq 0 ]]
