package client

// -build_flags=-mod=mod is required because of:
// https://github.com/golang/mock#reflect-vendoring-error.prog.go

//go:generate mockgen -build_flags=-mod=mod -package=client -destination mock_client.go sigs.k8s.io/controller-runtime/pkg/client Client,SubResourceClient
