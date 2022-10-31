# grpc-go-timeout

## per-request server-side timeout

- [Ignored](./main_test.go#L20)
- [Delt](./main_test.go#58)

It's the developer's responsibility to deal with the server-side timeout. If you ignore the context's deadline, `go` won't complain about it.

## client-side timeout

- [Client-side timeout](./main_test.go#L99)
- [Set with metadata](./main_test.go#129)

Client-side timeout won't be ignored, but the developer can choose to time out early.

Client-side timeout can't be set with gRPC metadata, becasue `grpc-timeout` information is within the HTTP2 [Request-Headers frame](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests).
