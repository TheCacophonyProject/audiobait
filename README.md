# audiobait

[![Build Status](https://api.travis-ci.com/TheCacophonyProject/audiobait.svg?branch=master)](https://travis-ci.com/TheCacophonyProject/audiobait)

`audiobait` plays audio files on a configurable schedule. The audio
files played are reported as events to the Cacophony Project [Events
API](https://api.cacophony.org.nz/#api-Events-Add_Event) (via
[event-reporter](https://github.com/TheCacophonyProject/event-reporter)).

This software is licensed under the GNU General Public License v3.0.

## Releases

This software uses the [GoReleaser](https://goreleaser.com) tool to
automate releases. To produce a release:

* Ensure that the `GITHUB_TOKEN` environment variable is set with a
  Github personal access token which allows access to the Cacophony
  Project repositories.
* Tag the release with an annotated tag. For example:
  `git tag -a "v1.4" -m "1.4 release"`
* Push the tag to Github: `git push --tags origin`
* Run `goreleaser --rm-dist`

The configuration for GoReleaser can be found in `.goreleaser.yml`.
