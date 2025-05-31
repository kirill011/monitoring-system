proto_gen:
	protoc --proto_path=proto/messages --go_out=proto/messages --go_opt=paths=source_relative messages.proto
	protoc --proto_path=proto/devices --go_out=proto/devices --go_opt=paths=source_relative devices.proto

start_service_rebuild:
	docker compose up --build notifycation-service

start_service:
	docker compose up