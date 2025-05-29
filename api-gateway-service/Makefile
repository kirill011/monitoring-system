proto_gen:
	protoc --proto_path=proto/api-gateway/users --go_out=proto/api-gateway/users --go_opt=paths=source_relative users.proto
	protoc --proto_path=proto/api-gateway/devices --go_out=proto/api-gateway/devices --go_opt=paths=source_relative devices.proto


start_service_rebuild:
	docker compose up --build api-gateway-service

start_service:
	docker compose up