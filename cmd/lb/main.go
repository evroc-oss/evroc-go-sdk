// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Mini CLI for exercising load balancer CRUD against any environment.
//
// Usage:
//
//	go run ./cmd/lb <resource> <operation> [args...]
//
// Resources: lb, pool, svc, route
//
// Operations:
//
//	list                         List all resources
//	get    <name>                Get a single resource
//	create <name> [flags]        Create a resource
//	delete <name>                Delete a resource
//	patch  <name> <json>         Patch a resource with raw JSON
//
// Create flags (vary by resource):
//
//	lb     --ip-ref <ref> --listener-port <port> --route-ref <ref>
//	pool   --backend-refs <ref,ref,...>
//	svc    --port <port> --pool-ref <ref> [--proxy-protocol]
//	route  --svc-ref <ref>
//
// Examples:
//
//	go run ./cmd/lb pool list
//	go run ./cmd/lb pool create my-pool --backend-refs vm-1,vm-2
//	go run ./cmd/lb svc  create my-svc --port 8080 --pool-ref "$(go run ./cmd/lb pool ref my-pool)"
//	go run ./cmd/lb lb   get my-lb
//	go run ./cmd/lb pool delete my-pool
//	go run ./cmd/lb pool ref my-pool         # prints the fully-qualified reference
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/loadbalancer"
	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	resource := os.Args[1]
	operation := os.Args[2]
	args := os.Args[3:]

	ctx := context.Background()
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	lb := client.LoadBalancer()

	switch resource {
	case "lb":
		runLB(ctx, lb, operation, args)
	case "pool":
		runPool(ctx, lb, operation, args)
	case "svc":
		runSvc(ctx, lb, operation, args)
	case "route":
		runRoute(ctx, lb, operation, args)
	default:
		fatalf("unknown resource %q (use: lb, pool, svc, route)", resource)
	}
}

// --- LoadBalancer ---

func runLB(ctx context.Context, lb *loadbalancer.Client, op string, args []string) {
	switch op {
	case "list":
		res, err := lb.LoadBalancers().List(ctx)
		must(err, "list load balancers")
		printJSON(res.Items)

	case "get":
		requireArgs(args, 1, "get <name>")
		res, err := lb.LoadBalancers().Get(ctx, args[0])
		must(err, "get load balancer")
		printJSON(res)

	case "create":
		requireArgs(args, 1, "create <name> --ip-ref <ref> [--listener-port <port> --route-ref <ref>]")
		name := args[0]
		flags := parseFlags(args[1:])

		ipRef := flags.require("ip-ref")
		builder := loadbalancer.NewLoadBalancerBuilder(name).
			WithPublicIPRef(ipRef)

		if port := flags.getInt("listener-port", 0); port > 0 {
			listenerName := "default"
			item := lbtypes.LoadbalancerSpecListenersItem{
				Name:     &listenerName,
				Port:     int32(port),
				Protocol: lbtypes.TCP,
			}
			if routeRef := flags.get("route-ref"); routeRef != "" {
				refs := []string{routeRef}
				item.RouteRefs = &refs
			}
			builder = builder.WithListener(item)
		}

		res, err := builder.Create(ctx, lb.LoadBalancers())
		must(err, "create load balancer")
		printJSON(res)

	case "delete":
		requireArgs(args, 1, "delete <name>")
		must(lb.LoadBalancers().Delete(ctx, args[0]), "delete load balancer")
		fmt.Println("deleted")

	case "patch":
		requireArgs(args, 2, "patch <name> <json>")
		var patch map[string]interface{}
		must(json.Unmarshal([]byte(args[1]), &patch), "parse patch JSON")
		res, err := lb.LoadBalancers().Patch(ctx, args[0], patch)
		must(err, "patch load balancer")
		printJSON(res)

	case "ref":
		requireArgs(args, 1, "ref <name>")
		fmt.Println(lb.LoadBalancerRef(args[0]))

	default:
		fatalf("unknown operation %q for lb", op)
	}
}

// --- BackendPool ---

func runPool(ctx context.Context, lb *loadbalancer.Client, op string, args []string) {
	switch op {
	case "list":
		res, err := lb.BackendPools().List(ctx)
		must(err, "list backend pools")
		printJSON(res.Items)

	case "get":
		requireArgs(args, 1, "get <name>")
		res, err := lb.BackendPools().Get(ctx, args[0])
		must(err, "get backend pool")
		printJSON(res)

	case "create":
		requireArgs(args, 1, "create <name> [--backend-refs <ref,ref,...>]")
		name := args[0]
		flags := parseFlags(args[1:])

		builder := loadbalancer.NewBackendPoolBuilder(name)
		if refs := flags.get("backend-refs"); refs != "" {
			builder = builder.WithBackendRefs(strings.Split(refs, ","))
		}

		res, err := builder.Create(ctx, lb.BackendPools())
		must(err, "create backend pool")
		printJSON(res)

	case "delete":
		requireArgs(args, 1, "delete <name>")
		must(lb.BackendPools().Delete(ctx, args[0]), "delete backend pool")
		fmt.Println("deleted")

	case "patch":
		requireArgs(args, 2, "patch <name> <json>")
		var patch map[string]interface{}
		must(json.Unmarshal([]byte(args[1]), &patch), "parse patch JSON")
		res, err := lb.BackendPools().Patch(ctx, args[0], patch)
		must(err, "patch backend pool")
		printJSON(res)

	case "ref":
		requireArgs(args, 1, "ref <name>")
		fmt.Println(lb.BackendPoolRef(args[0]))

	default:
		fatalf("unknown operation %q for pool", op)
	}
}

// --- BackendService ---

func runSvc(ctx context.Context, lb *loadbalancer.Client, op string, args []string) {
	switch op {
	case "list":
		res, err := lb.BackendServices().List(ctx)
		must(err, "list backend services")
		printJSON(res.Items)

	case "get":
		requireArgs(args, 1, "get <name>")
		res, err := lb.BackendServices().Get(ctx, args[0])
		must(err, "get backend service")
		printJSON(res)

	case "create":
		requireArgs(args, 1, "create <name> --port <port> [--pool-ref <ref>] [--proxy-protocol]")
		name := args[0]
		flags := parseFlags(args[1:])

		port := flags.requireInt("port")
		builder := loadbalancer.NewBackendServiceBuilder(name).
			WithPort(int32(port))

		if poolRef := flags.get("pool-ref"); poolRef != "" {
			builder = builder.WithBackendPoolRef(poolRef)
		}
		if flags.getBool("proxy-protocol") {
			builder = builder.WithProxyProtocol(true)
		}

		res, err := builder.Create(ctx, lb.BackendServices())
		must(err, "create backend service")
		printJSON(res)

	case "delete":
		requireArgs(args, 1, "delete <name>")
		must(lb.BackendServices().Delete(ctx, args[0]), "delete backend service")
		fmt.Println("deleted")

	case "patch":
		requireArgs(args, 2, "patch <name> <json>")
		var patch map[string]interface{}
		must(json.Unmarshal([]byte(args[1]), &patch), "parse patch JSON")
		res, err := lb.BackendServices().Patch(ctx, args[0], patch)
		must(err, "patch backend service")
		printJSON(res)

	case "ref":
		requireArgs(args, 1, "ref <name>")
		fmt.Println(lb.BackendServiceRef(args[0]))

	default:
		fatalf("unknown operation %q for svc", op)
	}
}

// --- L4Route ---

func runRoute(ctx context.Context, lb *loadbalancer.Client, op string, args []string) {
	switch op {
	case "list":
		res, err := lb.L4Routes().List(ctx)
		must(err, "list L4 routes")
		printJSON(res.Items)

	case "get":
		requireArgs(args, 1, "get <name>")
		res, err := lb.L4Routes().Get(ctx, args[0])
		must(err, "get L4 route")
		printJSON(res)

	case "create":
		requireArgs(args, 1, "create <name> --svc-ref <ref>")
		name := args[0]
		flags := parseFlags(args[1:])

		svcRef := flags.require("svc-ref")
		res, err := loadbalancer.NewL4RouteBuilder(name).
			WithBackendServiceRef(svcRef).
			Create(ctx, lb.L4Routes())
		must(err, "create L4 route")
		printJSON(res)

	case "delete":
		requireArgs(args, 1, "delete <name>")
		must(lb.L4Routes().Delete(ctx, args[0]), "delete L4 route")
		fmt.Println("deleted")

	case "patch":
		requireArgs(args, 2, "patch <name> <json>")
		var patch map[string]interface{}
		must(json.Unmarshal([]byte(args[1]), &patch), "parse patch JSON")
		res, err := lb.L4Routes().Patch(ctx, args[0], patch)
		must(err, "patch L4 route")
		printJSON(res)

	case "ref":
		requireArgs(args, 1, "ref <name>")
		fmt.Println(lb.L4RouteRef(args[0]))

	default:
		fatalf("unknown operation %q for route", op)
	}
}

// --- Helpers ---

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Fatalf("json encode: %v", err)
	}
}

func must(err error, action string) {
	if err != nil {
		log.Fatalf("failed to %s: %v", action, err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func requireArgs(args []string, n int, usageHint string) {
	if len(args) < n {
		fatalf("usage: go run ./cmd/lb <resource> %s", usageHint)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: go run ./cmd/lb <resource> <operation> [args...]

Resources: lb, pool, svc, route

Operations:
  list                    List all resources
  get    <name>           Get a single resource
  create <name> [flags]   Create a resource
  delete <name>           Delete a resource
  patch  <name> <json>    Patch with raw JSON
  ref    <name>           Print fully-qualified reference

Run with EVROC_SDK_DEBUG=1 for request/response logging.`)
	os.Exit(1)
}

type flagSet map[string]string

func parseFlags(args []string) flagSet {
	fs := flagSet{}
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			continue
		}
		key := strings.TrimPrefix(args[i], "--")
		if key == "proxy-protocol" {
			fs[key] = "true"
			continue
		}
		if i+1 >= len(args) {
			fatalf("flag --%s requires a value", key)
		}
		i++
		fs[key] = args[i]
	}
	return fs
}

func (f flagSet) get(key string) string       { return f[key] }
func (f flagSet) getBool(key string) bool      { return f[key] == "true" }

func (f flagSet) require(key string) string {
	v := f[key]
	if v == "" {
		fatalf("required flag --%s not provided", key)
	}
	return v
}

func (f flagSet) getInt(key string, defaultVal int) int {
	v := f[key]
	if v == "" {
		return defaultVal
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		fatalf("flag --%s must be an integer: %v", key, err)
	}
	return n
}

func (f flagSet) requireInt(key string) int {
	n := f.getInt(key, -1)
	if n < 0 {
		fatalf("required flag --%s not provided", key)
	}
	return n
}
