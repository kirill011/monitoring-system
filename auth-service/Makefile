proto_gen:
	protoc --proto_path=proto/users --go_out=proto/users --go_opt=paths=source_relative users.proto
	protoc --proto_path=proto/api-gateway/users --go_out=proto/api-gateway/users --go_opt=paths=source_relative apiusers.proto

start_service_rebuild:
	docker compose up --build auth-service

start_service:
	docker compose up