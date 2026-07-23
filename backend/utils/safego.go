package utils

import (
	"log"
	"runtime/debug"
)

// RecoverPanic recovers a panicking goroutine and logs it (with a stack trace)
// instead of letting the panic unwind to the top of the goroutine and crash the
// entire process. Fiber's middleware recover only protects the request
// goroutine — any `go ...` spawned for fire-and-forget work (audit logging,
// document sync, notifications) runs outside it, so a single nil-deref or bad
// type assertion there would take down the server for all tenants.
//
// Use it as the FIRST deferred call inside any goroutine body:
//
//	go func() {
//		defer utils.RecoverPanic("notify-approval-required")
//		// ... work that might panic ...
//	}()
func RecoverPanic(label string) {
	if r := recover(); r != nil {
		log.Printf("[panic-recovered] goroutine %q: %v\n%s", label, r, debug.Stack())
	}
}
