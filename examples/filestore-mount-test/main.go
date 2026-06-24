// Quick filestore mount test — create, wait, mount, write, unmount, done.
// Run on the evroc VM that needs NFS access to the filestore.
//
//	go run main.go
//
// Leaves the filestore running so you can test manually:
//
//	sudo mount -t nfs4 -o vers=4.1 <endpoint>:/ /mnt/filestore
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/storage"
)

const (
	fileStoreName = "nfs-mount-test"
	zone          = "a"
	mountPoint    = "/tmp/nfs-mount-test"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("SDK client: %v", err)
	}

	// 1. Create or reuse filestore
	fmt.Printf("1. Creating filestore %s (zone %s)...\n", fileStoreName, zone)
	_, err = storage.NewFileStoreBuilder(fileStoreName, zone).
		Create(ctx, client.Storage().FileStores())
	if err != nil {
		fmt.Printf("   create: %v (may already exist, continuing)\n", err)
	} else {
		fmt.Println("   created")
	}

	// 2. Wait for available
	fmt.Println("2. Waiting for filestore to become available...")
	fs, err := client.Storage().FileStores().WaitForAvailable(ctx, fileStoreName, 5*time.Minute)
	if err != nil {
		log.Fatalf("   not available: %v", err)
	}
	endpoint := fs.Status.Nfs.Endpoint
	exportPath := fs.Status.Nfs.ExportPath
	fmt.Printf("   available: %s:%s\n", endpoint, exportPath)

	// 3. Mount
	nfsSource := fmt.Sprintf("%s:%s", endpoint, exportPath)
	fmt.Printf("3. Mounting %s to %s...\n", nfsSource, mountPoint)

	os.MkdirAll(mountPoint, 0o755)
	out, err := exec.CommandContext(ctx, "mount", "-t", "nfs4", "-o", "vers=4.1", nfsSource, mountPoint).CombinedOutput()
	if err != nil {
		log.Fatalf("   mount failed: %v\n%s", err, string(out))
	}
	fmt.Println("   mounted")

	// 4. Write a test file
	testFile := mountPoint + "/hello.txt"
	testData := fmt.Sprintf("written at %s from pid %d\n", time.Now().Format(time.RFC3339), os.Getpid())
	fmt.Printf("4. Writing %s...\n", testFile)
	if err := os.WriteFile(testFile, []byte(testData), 0o644); err != nil {
		log.Fatalf("   write failed: %v", err)
	}

	// 5. Read it back
	data, err := os.ReadFile(testFile)
	if err != nil {
		log.Fatalf("   read failed: %v", err)
	}
	fmt.Printf("   read back: %s", string(data))

	// 6. Unmount
	fmt.Println("5. Unmounting...")
	out, err = exec.Command("umount", mountPoint).CombinedOutput()
	if err != nil {
		fmt.Printf("   umount: %v (%s)\n", err, string(out))
	} else {
		fmt.Println("   unmounted")
	}

	fmt.Printf("\nFilestore %s is still running at %s:%s\n", fileStoreName, endpoint, exportPath)
	fmt.Printf("Manual test:  sudo mount -t nfs4 -o vers=4.1 %s:%s /mnt/test\n", endpoint, exportPath)
}
