// Command main runs https server.
//
// # Usage
//
//	go run main.go [flags] [path ...]
//
// The flags are:
//
//	-v
//		Prints build version. (@unimplemented)
//
// # Build and Deploy
//
// Pre-Build Commands:
//
//	$ templ generate
//	$ tailwindcss -i .\templates\css\globals.css -o .\static\css\style.css --minify
//
// Build Command:
//
//	$ go build -tags netgo -ldflags '-s -w' -o app
//
// # Develop
//
// Install dependencies
//
//	$ go install github.com/a-h/templ/cmd/templ@latest
//
// Development with hot-module reloading: https://github.com/cosmtrek/air
//
//	$ air init
//	$ air
//
// Install templ
//
//	$ go install github.com/a-h/templ/cmd/templ@latest
//	$ templ generate
//
// # Document using godoc
//
// Install with:
//
//	$ go install golang.org/x/tools/cmd/godoc@latest
//
// Usage:
//
//	$ godoc -http=localhost:6060
//
// Access it in your web browser by navigating to http://localhost:6060 (or whichever port you specified).
//
// Jump to project: http://localhost:6060/pkg/github.com/lloydlobo/go-headcount/
//
// # References
//
//   - https://github.dev/syarul/todomvc-go-templ-htmx-_hyperscript/blob/main/main.go
//   - https://github.com/google/exposure-notifications-server/blob/2041f77a0bda55a67214d23dc18f26b3ab895fd1/cmd/admin-console/main.go#L32
//
// # Errorlog
//
// 2024/02:
//
//	listen tcp :8000: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.
//
// What: While spamming POST "/contacts" -> should rate limit.
//
// Why: Seems like `air` dev tool error.
//
// 2024/02/07 17:17:13
//
//	http: superfluous response.WriteHeader call from github.com/lloydlobo/go-headcount/handlers.(*DefaultHandler).ContactPartialsHandler (default.go:130)
//
// What: when an invalid email, e.g. `hi@johndoe.c-om`; is received from new contact form.
//
// Why: guess that error is written and something else also writes status code after.
//
// Solved: method was not exiting early on error. added a return statement.
//
// 2024/02/11 14:37:20
//
//	error fetching user data from api: Get "https://jsonplaceholder.typicode.com/users": read tcp [2405:201:400f:40e2:2162:92d4:427a:2abb]:54885->[2606:4700:3031::ac43:a797]:443: wsarecv: An existing connection was forcibly closed by the remote host.
//
// What: during fast saving of files during hot reloading behavior via air.
//
// Why: fails to load remote fake data, and halts showing initial contacts in the frontend (although the app works.)
package main
