# cutroom

Video editing for AI agents. An agent can read text and look at images, but it can't watch video or
hear audio. cutroom turns footage into text and images the agent can reason over, turns its text
decisions into a frame-exact video, and then checks the render the same way, so the agent can prove
the result matches what it intended: **perceive → decide → act → verify**.

It is built for footage with speech: talking heads, tutorials, podcasts, interviews and shorts.
It handles cutting by quoted words, retake removal, pause tightening, punch-ins and vertical
reframing, text, image and B-roll overlays, burned-in and SRT/VTT captions, a separate audio
recorder, loudness, and export to other editors (OpenTimelineIO, CMX 3600 EDL). Sources are never
modified, cuts land on whole frames and audio samples, and video is encoded once.

## Install

cutroom needs ffmpeg, whisper.cpp and a whisper speech model; `cutroom doctor` (below) walks you
through them.

Install the latest release (macOS and Linux, Apple silicon / arm64 and x86-64):

```bash
curl -fsSL https://raw.githubusercontent.com/NWylynko/cutroom/main/install.sh | sh
```

It puts `cutroom` in `~/.local/bin` (set `CUTROOM_INSTALL_DIR` to change that, `CUTROOM_VERSION=v1.0.0`
for a specific release) and checks the download against the release's checksums. You can also
download an archive yourself from [Releases](https://github.com/NWylynko/cutroom/releases).

### From source

You need Go 1.27 or newer.

```bash
go install github.com/NWylynko/cutroom/cmd/cutroom@latest
```

or from a clone:

```bash
git clone https://github.com/NWylynko/cutroom.git && cd cutroom
go install ./cmd/cutroom
```

`go install` fetches two Go modules (golang.org/x/image, x/text) the first time, so it needs the
network once. It puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`); make sure that folder
is on your `PATH`.

### Dependencies

However you installed it, let cutroom check the rest:

```bash
cutroom doctor
```

`doctor` lists each dependency and, for anything missing, the exact command that installs it. On
macOS that is `brew install ffmpeg whisper.cpp` and a model download into
`~/.cache/cutroom/models` (`ggml-base.en.bin`, 148 MB; `large-v3-turbo` is more accurate). Run
each fix it prints, then `cutroom doctor` again until it says `all dependencies ok`. To keep models
elsewhere, set `CUTROOM_MODELS` to their folder (or `CUTROOM_WHISPER_MODEL` to one file).

## Quick start

In a folder with a video of someone talking (here `talk.mp4`):

```bash
cutroom init talk.mp4              # a project here: probes the video, warns about anything unusual
cutroom transcribe                 # word-level transcript (cached)
cutroom transcript --suspects      # retakes, stray sounds, doubtful words, before you cut
cutroom edit tighten               # shorten long pauses
cutroom timeline                   # the edit, and each WARN to decide (see below)
cutroom timeline --apply retake-kept   # take a warning's suggested fix
cutroom render --preview           # fast, small render: out/preview.mp4
cutroom check out/preview.mp4      # verify it: frames, picture, sync, every word, loudness
cutroom render                     # the full-quality render: out/final.mp4
cutroom check                      # verify that too
cutroom captions                   # out/captions.srt in the edit's time
```

**Deciding warnings.** `timeline` lists each `WARN[code] at output <time>` with the edit that fixes
it. Decide every one, one at a time: `cutroom timeline --apply <code>@<time>` runs that edit (e.g.
cut the earlier take of a retake, or an off-topic "by the way" aside), and `cutroom timeline
--accept <code>@<time> --reason "…"` keeps it on purpose (an aside the video needs). The time picks
one when a code repeats; `timeline` prints "no problems found" when all are decided.

Every command prints a `next:` line with what to do next, `--help` for its options and `--json`
for the complete result. `cutroom status` says where a project stands at any point. Edits are
undoable (`cutroom edit undo`). `cutroom look` and `cutroom listen` write PNGs of frames and of
sound so an agent can see and hear what it is cutting: `cutroom look --cuts` shows every cut,
`cutroom look --output 0:03-0:06` the edit at those times, `cutroom listen 16.5-17.8` a stretch of
the source's sound.

## Using it with an AI agent

Point the agent at cutroom and tell it to start with the manual:

```bash
cutroom guide
```

The guide is written for agents: the workflow, how to edit by quoting words, how to verify a
render, and the pitfalls. `cutroom guide <topic>` opens one topic (`workflow`, `edit`, `overlays`,
`captions`, `shorts`, `render`, `verify`, `pitfalls`, `perceive`).

## Development

```bash
make check     # vet + unit tests (including a guard test for every lesson in docs/LESSONS.md)
make e2e       # builds test fixtures, then runs the CLI end to end
```

`docs/SPEC.md` is what cutroom is meant to be, `docs/LESSONS.md` the production lessons it is built
on, and `LOOP.md` how it is developed.

## License

Apache 2.0; see [LICENSE](LICENSE).
