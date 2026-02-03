// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	evroc "github.com/evroc-oss/evroc-go-sdk"
)

func main() {
	// Create context that cancels on SIGINT/SIGTERM (Ctrl-C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create client - config has project/region
	client, err := evroc.NewFromEnv(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// List virtual machines
	vms, err := client.Compute().VirtualMachines().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d VMs\n", len(vms.Items))

	// Other operations work the same way:
	// vm, err := client.Compute().VirtualMachines().Get(ctx, "my-vm")
	// created, err := client.Compute().VirtualMachines().Create(ctx, vmSpec)
	// err := client.Compute().VirtualMachines().Delete(ctx, "my-vm")
}
