name=echo
pbname=$(name)pb
src_dir=proto/$(name)/v1
dest_dir=internal/$(pbname)
swagger_dir=gen/swagger

all: $(pbname)

$(pbname): $(src_dir)/$(name).proto
	@mkdir -p $(dest_dir)
	@mkdir -p $(swagger_dir)
	@protoc -I$(src_dir) \
		-I/usr/include/google/protobuf \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out $(dest_dir) \
		--go_opt paths=source_relative \
		--go-grpc_out $(dest_dir) \
		--go-grpc_opt paths=source_relative \
		--grpc-gateway_out $(dest_dir) \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=false \
		--grpc-gateway_opt allow_patch_feature=true \
		--swagger_out $(swagger_dir) \
		--swagger_opt logtostderr=true \
		$<

clean:
	@rm -rf $(dest_dir)
	@if [ -d $$(dirname "$(dest_dir)") ]; then rmdir $$(dirname "$(dest_dir)") --ignore-fail-on-non-empty; fi
	@rm -rf $(swagger_dir)
	@if [ -d $$(dirname "$(swagger_dir)") ]; then rmdir $$(dirname "$(swagger_dir)") --ignore-fail-on-non-empty; fi

