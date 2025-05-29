proto_gen:
	protoc --proto_path=proto/devices --go_out=proto/devices --go_opt=paths=source_relative devices.proto
	protoc --proto_path=proto/api-gateway/devices --go_out=proto/api-gateway/devices --go_opt=paths=source_relative apidevices.proto

start_service_rebuild:
	docker compose up --build device-management-service

start_service:
	docker compose up