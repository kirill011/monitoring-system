proto_gen:
	protoc --proto_path=proto/devices --go_out=proto/devices --go_opt=paths=source_relative devices.proto
	protoc --proto_path=proto/messages --go_out=proto/messages --go_opt=paths=source_relative messages.proto
	protoc --proto_path=proto/notification --go_out=proto/notification --go_opt=paths=source_relative notification.proto
	protoc --proto_path=proto/api-gateway/tags --go_out=proto/api-gateway/tags --go_opt=paths=source_relative apitags.proto

start_service_rebuild:
	docker compose up --build data-processing-service

start_service:
	docker compose up