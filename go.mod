module github.com/duythinht/shout

go 1.21

require (
	github.com/kkdai/youtube/v2 v2.8.3
	github.com/slack-go/slack v0.12.2
	golang.org/x/net v0.17.0
)

require github.com/go-chi/chi/v5 v5.0.8

require (
	github.com/bitly/go-simplejson v0.5.1 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/dmulholl/mp3lib v1.0.0
	github.com/dop251/goja v0.0.0-20231027120936-b396bb4c349d // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/google/pprof v0.0.0-20231101202521-4ca4178f5c7a // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/kkdai/youtube/v2 => github.com/ppalone/youtube/v2 v2.0.0-20240307204930-212376cf3354
