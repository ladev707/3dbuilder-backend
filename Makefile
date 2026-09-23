# Learning note: .PHONY marks targets that are command names, not files.
# Without it, a file named "routes" in this folder would stop `make routes`
# from running.
.PHONY: debug routes

debug:
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient --log ./cmd/api

routes:
	go run ./cmd/console routes:show
